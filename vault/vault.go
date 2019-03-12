package vault

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/hashicorp/vault/api"
	"github.com/zipcar/bosh-vault/config"
	"github.com/zipcar/bosh-vault/logger"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const DefaultVaultRenewalIntervalSeconds = 3600

func GetVault(vaultConfig config.VaultConfiguration) (Vault, error) {
	var vault Vault

	if vaultConfig.RenewalInterval == 0 {
		vaultConfig.RenewalInterval = DefaultVaultRenewalIntervalSeconds
	}

	if vaultConfig.Mount == "" {
		vaultConfig.Mount = config.DefaultVaultMount
	}

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
			logger.Log.Fatalf("Failed to append %q to RootCAs: %v", vaultConfig.Ca, err)
		}

		if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
			logger.Log.Debug("No certs appended, using system certs only")
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

	ticker := time.NewTicker(time.Duration(vaultConfig.RenewalInterval) * time.Second)
	go func() {
		for _ = range ticker.C {
			_, err := vault.Client.Auth().Token().RenewSelf(vaultConfig.RenewalInterval)
			if err != nil {
				logger.Log.Errorf("Problem renewing token for %s, will try again in %d seconds", vaultConfig.Address, vaultConfig.RenewalInterval)
			}
		}
	}()
	return vault, nil
}

type Vault struct {
	Client *api.Client
	Config config.VaultConfiguration
}

func (v *Vault) exists(path string) bool {
	r := v.Client.NewRequest("GET", path)

	resp, _ := v.Client.RawRequest(r)
	if resp != nil {
		defer resp.Body.Close()
	}

	if resp != nil && resp.StatusCode == 200 {
		return true
	}

	return false
}

func (v *Vault) ExistsV1(name string) bool {
	path := fmt.Sprintf("/v1/%s%s", v.Config.Mount, v.sanitizeName(name))
	return v.exists(path)
}

func (v *Vault) Exists(name string) bool {
	path := fmt.Sprintf("/v1/%s", v.parseDataPath(name))
	return v.exists(path)
}

func (v *Vault) Get(name string, params map[string]string) (map[string]interface{}, error) {
	path := v.parseDataPath(name)
	vaultReply, err := kvReadRequest(v.Client, path, params)
	if err != nil {
		return nil, err
	}
	if vaultReply == nil {
		return nil, errors.New("secret not found")
	}

	return vaultReply.Data, nil
}

func (v *Vault) Delete(name string) error {
	_, err := v.Client.Logical().Delete(v.parseDataPath(name))
	return err
}

func (v *Vault) GetMetadata(name string) (map[string]interface{}, error) {
	metadataPath := v.parseMetaDataPath(name)
	metadata, err := v.Client.Logical().Read(metadataPath)
	if err != nil {
		return nil, err
	}

	if metadata == nil {
		return nil, errors.New(fmt.Sprintf("no metadata available for %s", name))
	}

	return metadata.Data, nil
}

func (v *Vault) Set(name string, value interface{}) (map[string]interface{}, error) {
	path := v.parseDataPath(name)
	response, err := v.Client.Logical().Write(path, map[string]interface{}{
		"data":    value,
		"options": map[string]interface{}{},
	})
	if err != nil {
		return nil, err
	}
	return response.Data, nil
}

func (v *Vault) Healthy() bool {
	healthResponse, err := v.Client.Sys().Health()
	if err != nil {
		logger.Log.Errorf("problem checking health of vault %s: %s", v.Config.Address, err)
		return false
	}
	return healthResponse.Initialized && !healthResponse.Sealed
}

func (v *Vault) sanitizeName(name string) string {
	// name should not have spaces
	name = strings.Replace(name, " ", "", -1)
	// names in this context should start with a /
	if !strings.HasPrefix(name, "/") {
		name = "/" + name
	}
	return name
}

func (v *Vault) parseDataPath(name string) string {
	return fmt.Sprintf("%s/data%s", v.Config.Mount, v.sanitizeName(name))
}

func (v *Vault) parseMetaDataPath(name string) string {
	return fmt.Sprintf("%s/metadata%s", v.Config.Mount, v.sanitizeName(name))
}
