package store

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

type SecretResponse struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
	Id    string      `json:"id"`
}

var vaultRouter redirectEngine
var defaultVault vault

func Initialize(bvConfig config.Configuration) {
	var err error
	defaultVault, err = getVault(bvConfig.Vault)
	if err != nil {
		// if we can't connect to the default backend that's a fatal error
		logger.Log.Fatalf("could not communicate default backend Vault server at %s, %s", bvConfig.Vault.Address, err)
	}

	for redirectConfigIndex, redirectConfiguration := range bvConfig.Redirects {
		vault, err := getVault(redirectConfiguration.Vault)
		vaultRouter.Vaults = append(vaultRouter.Vaults, vault)
		if err != nil {
			logger.Log.Errorf("Error establishing a connection to %s for redirects", redirectConfiguration.Vault.Address)
		}
		for _, rules := range redirectConfiguration.Rules {
			var redirect redirect
			redirect.Ref = rules.Ref
			redirect.Redirect = rules.Redirect
			redirect.Vault = &vaultRouter.Vaults[redirectConfigIndex]
			vaultRouter.Redirects = append(vaultRouter.Redirects, redirect)
		}
	}
}

func GetLatestByName(name string) (SecretResponse, error) {
	vault := vaultRouter.routeVault(name)
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

	vault := vaultRouter.routeVault(name)
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

	vault := vaultRouter.routeVault(decodedId.Name)
	secret, err := kvReadRequest(vault.Client, decodedId.Path, map[string]string{
		"version": fmt.Sprintf("%s", decodedId.Version),
	})
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

func DeleteByName(name string) error {
	vault := defaultVault
	_, err := vault.Client.Logical().Delete(vault.parseDataPath(name))
	if err != nil {
		logger.Log.Error(err)
	}
	return nil
}

func SetSecret(name string, value interface{}) (string, error) {
	secretValue := map[string]interface{}{
		"data":    value,
		"options": map[string]interface{}{},
	}
	vault := defaultVault
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

type vault struct {
	Client *api.Client
	Config config.VaultConfiguration
}

func (v *vault) parseDataPath(name string) string {
	// todo implement redirect path change
	return fmt.Sprintf("%s/data%s", v.Config.Prefix, v.nameToPath(name))
}

func (v *vault) parseMetaDataPath(name string) string {
	// todo implement redirect path change
	return fmt.Sprintf("%s/metadata%s", v.Config.Prefix, v.nameToPath(name))
}

func (v *vault) nameToPath(name string) string {
	// todo: spaces are a problem for the network path but full url encoding is a problem for Vault...
	//  see if there are other characters and solve this encoding issue
	return strings.Replace(name, " ", "", -1)
}

func getVault(vaultConfig config.VaultConfiguration) (vault, error) {
	var vault vault
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

	// Don't need to pass a timeout value in this config since it's a shortcut for setting the HTTP client one (above)
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

type redirect struct {
	Ref      string
	Redirect string
	Vault    *vault
}

type redirectEngine struct {
	Redirects []redirect
	Vaults    []vault
}

func (r *redirectEngine) routeVault(ref string) *vault {
	for _, rule := range r.Redirects {
		if ref == rule.Ref {
			return rule.Vault
		}
	}
	return &defaultVault
}
