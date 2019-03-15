package vault_test

import (
	"github.com/hashicorp/vault-plugin-secrets-kv"
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/http"
	"github.com/hashicorp/vault/logical"
	hashiVault "github.com/hashicorp/vault/vault"
	"github.com/zipcar/bosh-vault/config"
	"github.com/zipcar/bosh-vault/store"
	"github.com/zipcar/bosh-vault/vault"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var healthyVault vault.Vault

var seedForKnownMetaDataTest = store.VaultTestData{
	Path: "some_entry",
	Value: map[string]interface{}{
		"value": "who cares",
	},
}

var seedData = []store.VaultTestData{
	store.VaultTestData{
		Path: "some_password",
		Value: map[string]interface{}{
			"value": "$$$$$wakawakawaka$$$$$",
		},
		Seed: false, // A test expects to write this in for the first time
	},
	store.VaultTestData{
		Path: "some_value",
		Value: map[string]interface{}{
			"value": "theBestValue",
		},
		Seed: true,
	},
	store.VaultTestData{
		Path: "some_value2",
		Value: map[string]interface{}{
			"value":       "theBestValue",
			"with_other":  "settings too",
			"with_nested": "{\"a\":\"b\"}",
		},
		Seed: true,
	},
	store.VaultTestData{
		Path: "some_unicode",
		Value: map[string]interface{}{
			"value": "Ω≈ç√∫˜µ≤≥÷",
		},
	},
}

func TestVault(t *testing.T) {
	RegisterFailHandler(Fail)

	var err error
	core, _, token := hashiVault.TestCoreUnsealedWithConfig(t, &hashiVault.CoreConfig{
		LogicalBackends: map[string]logical.Factory{
			"kv": kv.Factory,
		},
	})

	listener, addr := http.TestServer(t, core)
	http.TestServerAuth(t, addr, token)

	vc, err := vault.GetVault(config.VaultConfiguration{
		Address: addr,
		Token:   token,
		Mount:   "config-server",
	})

	err = vc.Client.Sys().Mount("config-server", &api.MountInput{
		Type:        "kv-v2",
		Description: "some config server stuff",
	})

	Expect(err).NotTo(HaveOccurred())
	healthyVault = vc

	RunSpecs(t, "Vault Suite")

	listener.Close()
}
