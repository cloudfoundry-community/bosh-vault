package types

import (
	"github.com/sethvargo/go-password/password"
	"github.com/zipcar/bosh-vault/vault"
)

const PasswordType = "password"
const PasswordDefaultLength = 40
const PasswordDefaultSymbols = 0
const PasswordDefaultNumbers = 10
const PasswordDefaultNoUppercase = true
const PasswordDefaultAllowRepeat = true

type PasswordRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type PasswordRecord string

func (record PasswordRecord) Store(name string) (CredentialResponse, error) {
	var respObj PasswordResponse
	id, err := vault.StoreSecret(name, map[string]interface{}{
		"value": record,
		"type":  PasswordType,
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

func ParseVaultDataAsPassword(vaultData *vault.SecretResponse) *vault.SecretResponse {
	vaultData.Value = vaultData.Value.(map[string]interface{})["value"].(string)
	return vaultData
}

func (r *PasswordRequest) Validate() bool {
	return r.Type == PasswordType
}

func (r *PasswordRequest) Generate() (CredentialResponse, error) {
	var respObj PasswordResponse
	//todo: accept options and pass them through
	passValue, err := password.Generate(PasswordDefaultLength, PasswordDefaultNumbers, PasswordDefaultSymbols, PasswordDefaultNoUppercase, PasswordDefaultAllowRepeat)
	if err != nil {
		return respObj, err
	}

	return PasswordRecord(passValue).Store(r.Name)
}

func (r *PasswordRequest) CredentialType() string {
	return r.Type
}

type PasswordResponse struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}
