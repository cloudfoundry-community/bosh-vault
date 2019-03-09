package store

import (
	"encoding/json"
	"github.com/zipcar/bosh-vault/logger"
	"github.com/zipcar/bosh-vault/secret"
	"github.com/zipcar/bosh-vault/vault"
)

const v1Redirect = "v1"
const dynamicRedirect = "dynamic"

type Rule struct {
	Ref      string
	Redirect string
	Type     string
	Vault    *vault.Vault
}

type RedirectStore struct {
	Rules        []Rule
	Vaults       []vault.Vault
	DefaultVault vault.Vault
}

func (rs *RedirectStore) refRule(ref string) (bool, Rule) {
	for _, rule := range rs.Rules {
		if ref == rule.Ref {
			if rule.Vault.Healthy() {
				return true, rule
			} else {
				break
			}
		}
	}
	return false, Rule{}
}

func (rs *RedirectStore) normalizeSecret(s secret.Secret, originalName string) (secret.Secret, error) {
	decodedSecretId, err := DecodeId(s.Id)
	if err != nil {
		return s, err
	}

	normalizedId, err := EncodeId(VersionedSecretMetaData{
		Name:    originalName,
		Version: decodedSecretId.Version,
	})
	if err != nil {
		return s, err
	}

	s.Name = originalName
	s.Id = normalizedId

	return s, nil
}

func (rs *RedirectStore) Healthy() bool {
	return rs.DefaultVault.Healthy()
}

func (rs *RedirectStore) Exists(name string) bool {
	// assumption is that EXISTENCE always refers to the expected location and default Vault
	return rs.DefaultVault.Exists(name)
}

func (rs *RedirectStore) GetLatestByName(name string) (secret.Secret, error) {
	secrets, err := rs.GetByName(name)
	if err != nil {
		return secret.Secret{}, err
	}
	return secrets[0], nil
}

func (rs *RedirectStore) GetByName(name string) (secrets []secret.Secret, err error) {
	originalName := name

	redirected, rule := rs.refRule(name)
	if redirected {
		switch rule.Type {
		case v1Redirect, dynamicRedirect:
			vaultResponse, err := rule.Vault.Client.Logical().Read(rule.Redirect)
			if err != nil {
				logger.Log.Errorf("Problem handling redirect rule type:%s redirect: %s -> %s", rule.Type, name, rule.Redirect)
				return secrets, err
			}
			secretRequest := VersionedSecretMetaData{
				Name:    rule.Redirect,
				Version: json.Number("0"), // always fetch latest from redirect Vault
			}
			id, _ := EncodeId(secretRequest)
			secrets = []secret.Secret{{
				Name:  rule.Redirect,
				Id:    id,
				Value: vaultResponse.Data,
			}}
		default:
			secrets, err = getByName(rule.Vault, rule.Redirect)
		}
	} else {
		secrets, err = getByName(&rs.DefaultVault, name)
	}

	if err != nil || !redirected {
		return secrets, err
	}

	for i, s := range secrets {
		secrets[i], _ = rs.normalizeSecret(s, originalName)
	}

	// secrets are meant to be returned by this end point in reverse order (newest first) so when we're persisting
	// we need to persist in the reverse order of that or things could break when doing a local fail over
	for i := len(secrets) - 1; i >= 0; i-- {
		_, err := setSecret(&rs.DefaultVault, secrets[i].Name, secrets[i].Value)
		if err != nil {
			logger.Log.Errorf("Unable to cache redirected secret %s version %d in the default Vault", secrets[i].Name, len(secrets)-i)
		}
	}

	return secrets, nil
}

func (rs *RedirectStore) GetById(id string) (s secret.Secret, err error) {
	originalId := id

	decodedId, err := DecodeId(id)
	if err != nil {
		return secret.Secret{}, err
	}

	redirected, rule := rs.refRule(decodedId.Name)
	if redirected {
		secretRequest := VersionedSecretMetaData{
			Name:    rule.Redirect,
			Version: json.Number("0"), // always fetch latest from redirect Vault
		}
		id, _ = EncodeId(secretRequest)
		switch rule.Type {
		case v1Redirect, dynamicRedirect:
			// dynamic and v1 redirects will always be asked for by name FIRST and
			// cached in the default Vault so get the cached value, redeploys will
			// ask for the variable by name again, thus regenerating it.
			return getById(&rs.DefaultVault, originalId)
		default:
			s, err = getById(rule.Vault, id)
		}
	} else {
		s, err = getById(&rs.DefaultVault, id)
	}

	if err != nil || !redirected {
		return s, err
	}

	_, err = setSecret(&rs.DefaultVault, decodedId.Name, s.Value)
	if err != nil {
		logger.Log.Errorf("Unable to cache redirected secret %s in the default Vault", decodedId.Name)
	}

	s.Id = originalId
	return rs.normalizeSecret(s, decodedId.Name)
}

func (rs *RedirectStore) Set(name string, value interface{}) (string, error) {
	return setSecret(&rs.DefaultVault, name, value)
}

func (rs *RedirectStore) DeleteByName(name string) error {
	return deleteByName(&rs.DefaultVault, name)
}
