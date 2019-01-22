package store

import (
	"github.com/zipcar/bosh-vault/secret"
	"github.com/zipcar/bosh-vault/vault"
)

type VaultStore struct {
	Vault vault.Vault
}

func (vs *VaultStore) Healthy() bool {
	return vs.Vault.Healthy()
}

func (vs *VaultStore) GetLatestByName(name string) (secret.Secret, error) {
	return getLatestByName(&vs.Vault, name)
}

func (vs *VaultStore) GetAllByName(name string) ([]secret.Secret, error) {
	return getAllByName(&vs.Vault, name)
}

func (vs *VaultStore) GetById(id string) (secret.Secret, error) {
	return getById(&vs.Vault, id)
}

func (vs *VaultStore) Set(name string, value interface{}) (string, error) {
	return setSecret(&vs.Vault, name, value)
}

func (vs *VaultStore) DeleteByName(name string) error {
	return deleteByName(&vs.Vault, name)
}
