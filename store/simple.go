package store

import (
	"github.com/zipcar/bosh-vault/secret"
	"github.com/zipcar/bosh-vault/vault"
)

type SimpleStore struct {
	Vault vault.Vault
}

func (vs *SimpleStore) Healthy() bool {
	return vs.Vault.Healthy()
}

func (vs *SimpleStore) GetByName(name string) ([]secret.Secret, error) {
	return getByName(&vs.Vault, name)
}

func (vs *SimpleStore) GetById(id string) (secret.Secret, error) {
	return getById(&vs.Vault, id)
}

func (vs *SimpleStore) Set(name string, value interface{}) (string, error) {
	return setSecret(&vs.Vault, name, value)
}

func (vs *SimpleStore) DeleteByName(name string) error {
	return deleteByName(&vs.Vault, name)
}
