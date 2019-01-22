package store

import (
	"encoding/json"
	"github.com/zipcar/bosh-vault/logger"
	"github.com/zipcar/bosh-vault/secret"
	"github.com/zipcar/bosh-vault/vault"
)

type redirect struct {
	Ref      string
	Redirect string
	Vault    *vault.Vault
}

type VaultRedirectStore struct {
	Redirects    []redirect
	Vaults       []vault.Vault
	DefaultVault vault.Vault
}

func (vrs *VaultRedirectStore) refRedirect(ref string) (bool, redirect) {
	for _, rule := range vrs.Redirects {
		if ref == rule.Ref {
			if rule.Vault.Healthy() {
				return true, rule
			} else {
				break
			}
		}
	}
	return false, redirect{}
}

func (vrs *VaultRedirectStore) normalizeSecret(s secret.Secret, originalName string) (secret.Secret, error) {
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

func (vrs *VaultRedirectStore) Healthy() bool {
	return vrs.DefaultVault.Healthy()
}

func (vrs *VaultRedirectStore) GetLatestByName(name string) (secret.Secret, error) {
	v := &vrs.DefaultVault
	originalName := name
	redirected, rule := vrs.refRedirect(name)
	if redirected {
		name = rule.Redirect
		v = rule.Vault
	}
	s, err := getLatestByName(v, name)
	if err != nil || !redirected {
		return s, err
	}

	// Persist to the default vault if redirected
	if redirected {
		_, err := setSecret(&vrs.DefaultVault, originalName, s.Value)
		if err != nil {
			logger.Log.Errorf("Unable to cache redirected secret %s in the default Vault", originalName)
		}
	}

	return vrs.normalizeSecret(s, originalName)

}

func (vrs *VaultRedirectStore) GetAllByName(name string) ([]secret.Secret, error) {
	v := &vrs.DefaultVault
	originalName := name

	redirected, rule := vrs.refRedirect(name)
	if redirected {
		name = rule.Redirect
		v = rule.Vault
	}
	secrets, err := getAllByName(v, name)
	if err != nil || !redirected {
		return secrets, err
	}

	for i, s := range secrets {
		secrets[i], _ = vrs.normalizeSecret(s, originalName)
	}

	// Persist to the default vault if redirected
	if redirected {
		// secrets are meant to be returned by this end point in reverse order (newest first) so when we're persisting
		// we need to persist in the reverse order of that or things could break when doing a local fail over
		for i := len(secrets) - 1; i >= 0; i-- {
			_, err := setSecret(&vrs.DefaultVault, secrets[i].Name, secrets[i].Value)
			if err != nil {
				logger.Log.Errorf("Unable to cache redirected secret %s version %d in the default Vault", secrets[i].Name, len(secrets)-i)
			}
		}
	}

	return secrets, nil
}

func (vrs *VaultRedirectStore) GetById(id string) (secret.Secret, error) {
	v := &vrs.DefaultVault
	originalId := id

	decodedId, err := DecodeId(id)
	if err != nil {
		return secret.Secret{}, err
	}

	redirected, rule := vrs.refRedirect(decodedId.Name)
	if redirected {
		secretRequest := VersionedSecretMetaData{
			Name:    rule.Redirect,
			Version: json.Number("0"), // redirects always fetch latest from redirect Vault
		}
		id, _ = EncodeId(secretRequest)
		v = rule.Vault
	}

	s, err := getById(v, id)
	if err != nil || !redirected {
		return s, err
	}

	// Persist to the default vault if redirected
	if redirected {
		_, err := setSecret(&vrs.DefaultVault, decodedId.Name, s.Value)
		if err != nil {
			logger.Log.Errorf("Unable to cache redirected secret %s in the default Vault", decodedId.Name)
		}
	}

	s.Id = originalId
	return vrs.normalizeSecret(s, decodedId.Name)
}

func (vrs *VaultRedirectStore) Set(name string, value interface{}) (string, error) {
	return setSecret(&vrs.DefaultVault, name, value)
}

func (vrs *VaultRedirectStore) DeleteByName(name string) error {
	return deleteByName(&vrs.DefaultVault, name)
}
