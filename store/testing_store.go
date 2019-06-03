package store

import (
	"github.com/cloudfoundry-community/bosh-vault/config"
	"github.com/cloudfoundry-community/bosh-vault/vault"
	"github.com/hashicorp/vault-plugin-secrets-kv"
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/http"
	"github.com/hashicorp/vault/logical"
	hashiVault "github.com/hashicorp/vault/vault"
	"net"
	"testing"
)

type VaultTestData struct {
	Path  string
	Value interface{}
	Seed  bool
	Id    string
}

func TestHealthySimpleStore(t *testing.T) (SimpleStore, net.Listener, error) {
	core, _, token := hashiVault.TestCoreUnsealedWithConfig(t, &hashiVault.CoreConfig{
		LogicalBackends: map[string]logical.Factory{
			"kv": kv.Factory,
		},
	})

	ln, addr := http.TestServer(t, core)
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
	vs := SimpleStore{
		Vault: vc,
	}
	return vs, ln, err
}

func TestSealedSimpleStore(t *testing.T) (SimpleStore, net.Listener, error) {
	core, _, token := hashiVault.TestCoreUnsealed(t)
	ln, addr := http.TestServer(t, core)
	http.TestServerAuth(t, addr, token)
	vc, err := vault.GetVault(config.VaultConfiguration{
		Address: addr,
		Token:   token,
	})
	_ = vc.Client.Sys().Seal()
	vs := SimpleStore{
		Vault: vc,
	}
	return vs, ln, err
}

func TestUninitializedVaultSimpleStore(t *testing.T) (SimpleStore, net.Listener, error) {
	core := hashiVault.TestCore(t)
	ln, addr := http.TestServer(t, core)
	vc, err := vault.GetVault(config.VaultConfiguration{
		Address: addr,
		Token:   "",
	})
	vs := SimpleStore{
		Vault: vc,
	}
	return vs, ln, err
}
