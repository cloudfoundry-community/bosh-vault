package types

import (
	"encoding/json"
	"fmt"
)

const RsaKeypairType = "rsa"

type RsaKeypairRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func (r *RsaKeypairRequest) Validate() bool {
	return r.Type == RsaKeypairType
}

func (r *RsaKeypairRequest) Generate() (GenericCredentialResponse, error) {
	var resp RsaKeypairResponse
	err := json.Unmarshal([]byte(fmt.Sprintf(`{
		"name": "%s",
		"id":   "1337",
		"value": {
			"public_key":  "This is totally a ssh pub key",
			"private_key": "This is totally a ssh private key."
		}
	}`, r.Name)), &resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (r *RsaKeypairRequest) CredentialType() string {
	return r.Type
}

type RsaKeypairValue struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

type RsaKeypairResponse struct {
	Name  string          `json:"name"`
	Id    string          `json:"id"`
	Value RsaKeypairValue `json:"value"`
}

func (res RsaKeypairResponse) JsonString() string {
	structBytes, err := json.Marshal(res)
	if err != nil {
		return "{}"
	}
	return fmt.Sprintf("%s", structBytes)
}
