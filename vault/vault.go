package vault

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/vault/api"
	"github.com/zipcar/bosh-vault/config"
	"github.com/zipcar/bosh-vault/logger"
	"strings"
	"time"
)

type SecretResponse struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
	Id    string      `json:"id"`
}

var Client *api.Client

func InitializeClient(vcfcsConfig config.Configuration) {
	var err error
	Client, err = api.NewClient(&api.Config{
		Address: vcfcsConfig.Vault.Address,
	})
	if err != nil {
		logger.Log.Fatalf("could not communicate with Vault server at %s, %s", vcfcsConfig.Vault.Address, err)
	}
	Client.SetToken(vcfcsConfig.Vault.Token)
	Client.SetClientTimeout(time.Duration(vcfcsConfig.Vault.Timeout) * time.Second)
}

func FetchSecretByName(name string) (SecretResponse, error) {
	fullPath := getFullPath(name)
	secretRequest := VersionedSecretMetaData{
		Name:    name,
		Path:    fullPath,
		Version: json.Number("0"), // version 0 will fetch latest which is the expected behavior when fetching by name
	}
	id, _ := EncodeId(secretRequest)
	return FetchSecretById(id)
}

func FetchSecretById(id string) (SecretResponse, error) {
	var response SecretResponse
	decodedId, err := DecodeId(id)
	if err != nil {
		return response, err
	}

	versionParam := map[string]string{
		"version": fmt.Sprintf("%s", decodedId.Version),
	}

	secret, err := kvReadRequest(Client, decodedId.Path, versionParam)
	if err != nil {
		return response, err
	}

	response = SecretResponse{
		Id:    id,
		Value: secret.Data["data"],
		Name:  decodedId.Name,
	}

	return response, nil
}

func DeleteSecretByName(name string) error {
	_, err := Client.Logical().Delete(getFullPath(name))
	if err != nil {
		logger.Log.Error(err)
	}
	return nil
}

func StoreSecret(name string, value interface{}) (string, error) {
	secretValue := map[string]interface{}{
		"data":    value,
		"options": map[string]interface{}{},
	}
	path := getFullPath(name)
	secret, err := Client.Logical().Write(path, secretValue)
	if err != nil {
		logger.Log.Error(err)
		return "", err
	}
	version, ok := secret.Data["version"].(json.Number)
	if !ok {
		logger.Log.Errorf("couldn't fetch secret version from data: %+v", secret.Data)
	}
	secretRecord := VersionedSecretMetaData{
		Version: version,
		Path:    path,
		Name:    name,
	}
	id, err := EncodeId(secretRecord)
	if err != nil {
		logger.Log.Error(err)
		return "", err
	}
	return id, nil
}

func getFullPath(name string) string {
	// todo: spaces are a problem for the network path but full url encoding is a problem for Vault... see if there are other characters and solve this encoding issue
	escapedName := strings.Replace(name, " ", "", -1)
	return fmt.Sprintf("secret/data%s", escapedName)
}
