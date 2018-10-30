package types

import (
	"github.com/sethvargo/go-password/password"
	"github.com/zipcar/vault-cfcs/vault"
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

func PasswordMarshalVaultData(password string) map[string]interface{} {
	return map[string]interface{}{
		"value": password,
		"type":  PasswordType,
	}
}

func PasswordUnmarshalVaultData(vaultData *vault.SecretResponse) *vault.SecretResponse {
	vaultData.Value = vaultData.Value.(map[string]interface{})["value"].(string)
	return vaultData
}

func (r *PasswordRequest) Validate() bool {
	return r.Type == PasswordType
}

func (r *PasswordRequest) Generate() (GenericCredentialResponse, error) {
	var respObj PasswordPostResponse
	//todo: accept options and pass them through
	passValue, err := password.Generate(PasswordDefaultLength, PasswordDefaultNumbers, PasswordDefaultSymbols, PasswordDefaultNoUppercase, PasswordDefaultAllowRepeat)
	if err != nil {
		return respObj, err
	}

	id, err := vault.StoreSecret(r.Name, PasswordMarshalVaultData(passValue))

	if err != nil {
		return respObj, err
	}

	respObj = PasswordPostResponse{
		Name:  r.Name,
		Id:    id,
		Value: passValue,
	}

	return respObj, nil
}

func (r *PasswordRequest) CredentialType() string {
	return r.Type
}

type PasswordPostResponse struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}
