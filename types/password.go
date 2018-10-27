package types

import (
	"encoding/json"
	"fmt"
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

	id, err := vault.StoreSecret(r.Name, passValue)
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

func (res PasswordPostResponse) JsonString() string {
	structBytes, err := json.Marshal(res)
	if err != nil {
		return "{}"
	}
	return fmt.Sprintf("%s", structBytes)
}
