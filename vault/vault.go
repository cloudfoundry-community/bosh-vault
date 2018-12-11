package vault

import (
	"encoding/json"
	"errors"
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

func InitializeClient(bvConfig config.Configuration) {
	var err error
	Client, err = api.NewClient(&api.Config{
		Address: bvConfig.Vault.Address,
	})
	if err != nil {
		logger.Log.Fatalf("could not communicate with Vault server at %s, %s", bvConfig.Vault.Address, err)
	}
	Client.SetToken(bvConfig.Vault.Token)
	Client.SetClientTimeout(time.Duration(bvConfig.Vault.Timeout) * time.Second)
}

func GetLatestByName(name string) (SecretResponse, error) {
	fullPath := parseDataPath(name)
	secretRequest := VersionedSecretMetaData{
		Name:    name,
		Path:    fullPath,
		Version: json.Number("0"),
	}
	id, _ := EncodeId(secretRequest)
	return GetById(id)
}

func GetAllByName(name string) ([]SecretResponse, error) {
	fullPath := parseDataPath(name)
	secretVersions := make([]SecretResponse, 0)

	metadata, err := kvGetMetadata(Client, name)
	if err != nil {
		return secretVersions, err
	}

	versionsRaw, ok := metadata.Data["versions"]
	if !ok || versionsRaw == nil {
		return secretVersions, errors.New(fmt.Sprintf("Could not get version information for %s", name))
	}

	versionCount := len(versionsRaw.(map[string]interface{}))

	for i := versionCount; i > 0; i-- {
		secretRequest := VersionedSecretMetaData{
			Name:    name,
			Path:    fullPath,
			Version: json.Number(fmt.Sprintf("%d", i)),
		}
		id, _ := EncodeId(secretRequest)
		secretResp, err := GetById(id)
		if err != nil {
			logger.Log.Errorf("Problem fetching secret: %+v", secretRequest)
		} else {
			secretVersions = append(secretVersions, secretResp)
		}
	}

	return secretVersions, err
}

func GetById(id string) (SecretResponse, error) {
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
	_, err := Client.Logical().Delete(parseDataPath(name))
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
	path := parseDataPath(name)
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

func NameToPath(name string) string {
	// todo: spaces are a problem for the network path but full url encoding is a problem for Vault... see if there are other characters and solve this encoding issue
	return strings.Replace(name, " ", "", -1)
}

func parseDataPath(name string) string {
	return fmt.Sprintf("secret/data%s", NameToPath(name))
}

func parseMetaDataPath(name string) string {
	return fmt.Sprintf("secret/metadata%s", NameToPath(name))
}
