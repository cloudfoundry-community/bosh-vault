package types

import (
	"github.com/sethvargo/go-password/password"
	"github.com/zipcar/bosh-vault/secret"
)

const PasswordType = "password"
const PasswordDefaultLength = 40
const PasswordDefaultSymbols = 0
const PasswordDefaultNumbers = 10
const PasswordDefaultNoUppercase = false
const PasswordDefaultAllowRepeat = true

type PasswordRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type PasswordRecord string

func (record PasswordRecord) Store(secretStore secret.Store, name string) (CredentialResponse, error) {
	var respObj PasswordResponse
	id, err := secretStore.Set(name, map[string]interface{}{
		"value": record,
	})

	if err != nil {
		return respObj, err
	}

	respObj = PasswordResponse{
		Name:  name,
		Id:    id,
		Value: string(record),
	}

	return respObj, nil
}

func (r *PasswordRequest) Validate() bool {
	return r.Type == PasswordType
}

func (r *PasswordRequest) Generate(secretStore secret.Store) (CredentialRecordInterface, error) {
	//todo: accept options and pass them through
	passValue, err := password.Generate(PasswordDefaultLength, PasswordDefaultNumbers, PasswordDefaultSymbols, PasswordDefaultNoUppercase, PasswordDefaultAllowRepeat)
	if err != nil {
		return nil, err
	}

	return PasswordRecord(passValue), nil
}

func (r *PasswordRequest) CredentialType() string {
	return r.Type
}

func (r *PasswordRequest) CredentialName() string {
	return r.Name
}

type PasswordResponse struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}
