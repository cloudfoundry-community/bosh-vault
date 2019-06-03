package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cloudfoundry-community/bosh-vault/config"
	"github.com/cloudfoundry-community/bosh-vault/logger"
	"github.com/cloudfoundry-community/bosh-vault/secret"
	"github.com/cloudfoundry-community/bosh-vault/vault"
	"strconv"
)

func GetStore(bvConfig config.Configuration) secret.Store {
	defaultVault, err := vault.GetVault(bvConfig.Vault)
	if err != nil {
		// if we can't connect to the default backend that's a fatal error
		logger.Log.Fatalf("could not communicate with default backend Vault server at %s, %s", bvConfig.Vault.Address, err)
	}

	if len(bvConfig.Redirects) > 0 {
		var store RedirectStore
		store.DefaultVault = defaultVault

		for redirectConfigIndex, redirectConfiguration := range bvConfig.Redirects {
			v, err := vault.GetVault(redirectConfiguration.Vault)
			store.Vaults = append(store.Vaults, v)
			if err != nil {
				logger.Log.Errorf("Error establishing a connection to %s for redirects", redirectConfiguration.Vault.Address)
			}

			for _, rules := range redirectConfiguration.Rules {
				var redirect Rule
				redirect.Ref = rules.Ref
				redirect.Redirect = rules.Redirect
				redirect.Type = redirectConfiguration.Type
				redirect.Vault = &store.Vaults[redirectConfigIndex]
				store.Rules = append(store.Rules, redirect)
			}
		}
		return &store
	} else {
		var store SimpleStore
		store.Vault = defaultVault
		return &store
	}
}

func getByName(v *vault.Vault, name string) ([]secret.Secret, error) {
	secretVersions := make([]secret.Secret, 0)

	metadata, err := v.GetMetadata(name)
	if err != nil {
		return secretVersions, err
	}

	versionsRaw, ok := metadata["versions"]
	if !ok || versionsRaw == nil {
		return secretVersions, errors.New(fmt.Sprintf("Could not get version information for %s", name))
	}

	versionCount := len(versionsRaw.(map[string]interface{}))

	for i := versionCount; i > 0; i-- {
		// Skip any destroyed secret versions
		// todo: determine if we should also hide "soft deleted" secrets
		if versionsRaw.(map[string]interface{})[strconv.Itoa(i)].(map[string]interface{})["destroyed"].(bool) == true {
			continue
		}
		secretRequest := VersionedSecretMetaData{
			Name:    name,
			Version: json.Number(fmt.Sprintf("%d", i)),
		}
		id, _ := EncodeId(secretRequest)
		secretResp, err := getById(v, id)
		if err != nil {
			logger.Log.Errorf("Problem fetching secret: %+v", secretRequest)
		} else {
			secretVersions = append(secretVersions, secretResp)
		}
	}

	return secretVersions, err
}

func getById(v *vault.Vault, id string) (secret.Secret, error) {
	var response secret.Secret
	decodedId, err := DecodeId(id)
	if err != nil {
		return response, errors.New("malformed or invalid id")
	}

	val, err := v.Get(decodedId.Name, map[string]string{
		"version": fmt.Sprintf("%s", decodedId.Version),
	})
	if err != nil {
		return response, err
	}

	response = secret.Secret{
		Id:    id,
		Value: val["data"],
		Name:  decodedId.Name,
	}

	return response, nil
}

func deleteByName(v *vault.Vault, name string) error {
	return v.Delete(name)
}

func setSecret(v *vault.Vault, name string, value interface{}) (string, error) {
	response, err := v.Set(name, value)
	if err != nil {
		logger.Log.Error(err)
		return "", err
	}
	version, ok := response["version"].(json.Number)
	if !ok {
		logger.Log.Errorf("couldn't fetch secret version from data: %+v", response)
	}
	secretRecord := VersionedSecretMetaData{
		Version: version,
		Name:    name,
	}
	id, err := EncodeId(secretRecord)
	if err != nil {
		logger.Log.Error(err)
		return "", err
	}
	return id, nil
}
