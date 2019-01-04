package vault

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/vault/api"
	"github.com/zipcar/bosh-vault/config"
	"github.com/zipcar/bosh-vault/logger"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

type Vault struct {
	Client *api.Client
	Config config.VaultConfiguration
}
type SecretResponse struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
	Id    string      `json:"id"`
}

type Redirect struct {
	Ref      string
	Redirect string
	Vault    *Vault
}

type RedirectEngine struct {
	Redirects []Redirect
	Vaults    []Vault
}

func (r *RedirectEngine) getVaultClientForRef(ref string) *Vault {
	for _, rule := range r.Redirects {
		if ref == rule.Ref {
			return rule.Vault
		}
	}
	return &defaultClient
}

func GetVault(method string, ref string) *Vault {
	switch method {
	case http.MethodGet:
		logger.Log.Debugf("Get redirect client for %s", ref)
		return redirectEngine.getVaultClientForRef(ref)
	default:
		return &defaultClient

	}
}

var redirectEngine RedirectEngine
var defaultClient Vault

func InitializeVault(vaultConfig config.VaultConfiguration) (Vault, error) {
	var vault Vault
	vault.Config = vaultConfig
	// Get the SystemCertPool, continue with an empty pool on error
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		logger.Log.Error("problem reading system , cert pool, if no UAA CA cert was passed in the config expect TLS errors")
		rootCAs = x509.NewCertPool()
	}

	if vaultConfig.Ca != "" {
		certs, err := ioutil.ReadFile(vaultConfig.Ca)
		if err != nil {
			log.Fatalf("Failed to append %q to RootCAs: %v", vaultConfig.Ca, err)
		}

		if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
			log.Println("No certs appended, using system certs only")
		}
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: vaultConfig.SkipVerify,
		RootCAs:            rootCAs,
	}

	// Setup a custom transport that trusts our UAA Ca as well as the system's trusted certs
	customTransport := &http.Transport{TLSClientConfig: tlsConfig}
	customHttpClient := &http.Client{
		Timeout:   time.Second * time.Duration(vaultConfig.Timeout),
		Transport: customTransport,
	}

	clientInstance, err := api.NewClient(&api.Config{
		Address:    vaultConfig.Address,
		HttpClient: customHttpClient,
	})

	if err != nil {
		logger.Log.Debugf("could not communicate with Vault server at %s, %s", vaultConfig.Address, err)
		return vault, err
	}

	clientInstance.SetToken(vaultConfig.Token)
	clientInstance.SetClientTimeout(time.Duration(vaultConfig.Timeout) * time.Second)

	vault.Client = clientInstance
	return vault, nil
}

func Initialize(bvConfig config.Configuration) {
	var err error
	defaultClient, err = InitializeVault(bvConfig.Vault)
	if err != nil {
		// if we can't connect to the default backend that's a fatal error
		logger.Log.Fatalf("could not communicate default backend Vault server at %s, %s", bvConfig.Vault.Address, err)
	}

	for redirectConfigIndex, redirectConfiguration := range bvConfig.Redirects {
		vault, err := InitializeVault(redirectConfiguration.Vault)
		redirectEngine.Vaults = append(redirectEngine.Vaults, vault)
		if err != nil {
			logger.Log.Errorf("Error establishing a connection to %s for redirects", redirectConfiguration.Vault.Address)
		}
		for _, rules := range redirectConfiguration.Rules {
			var redirect Redirect
			redirect.Ref = rules.Ref
			redirect.Redirect = rules.Redirect
			redirect.Vault = &redirectEngine.Vaults[redirectConfigIndex]
			redirectEngine.Redirects = append(redirectEngine.Redirects, redirect)
		}
	}
}

func GetLatestByName(name string) (SecretResponse, error) {
	vault := GetVault(http.MethodGet, name)
	fullPath := vault.parseDataPath(name)
	secretRequest := VersionedSecretMetaData{
		Name:    name,
		Path:    fullPath,
		Version: json.Number("0"),
	}
	id, _ := EncodeId(secretRequest)
	return GetById(id)
}

func GetAllByName(name string) ([]SecretResponse, error) {
	secretVersions := make([]SecretResponse, 0)

	vault := GetVault(http.MethodGet, name)
	fullPath := vault.parseDataPath(name)
	metadata, err := kvGetMetadata(vault, name)
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

	vault := GetVault(http.MethodGet, decodedId.Name)
	secret, err := kvReadRequest(vault.Client, decodedId.Path, getVersionParam(decodedId.Version))
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
	vault := GetVault(http.MethodDelete, name)
	_, err := vault.Client.Logical().Delete(vault.parseDataPath(name))
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
	vault := GetVault(http.MethodPost, name)
	path := vault.parseDataPath(name)
	secret, err := vault.Client.Logical().Write(path, secretValue)
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

func getVersionParam(version json.Number) map[string]string {
	return map[string]string{
		"version": fmt.Sprintf("%s", version),
	}
}

func (v *Vault) parseDataPath(name string) string {
	return fmt.Sprintf("%s/data%s", v.Config.Prefix, NameToPath(name))
}

func (v *Vault) parseMetaDataPath(name string) string {
	return fmt.Sprintf("%s/metadata%s", v.Config.Prefix, NameToPath(name))
}
