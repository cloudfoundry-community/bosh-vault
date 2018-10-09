package types

import (
	"encoding/json"
	"fmt"
)

const SshKeypairType = "ssh"

type SshKeypairRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func (r *SshKeypairRequest) Validate() bool {
	return r.Type == SshKeypairType
}

func (r *SshKeypairRequest) Generate() (GenericCredentialResponse, error) {
	var resp SshKeypairResponse
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

func (r *SshKeypairRequest) CredentialType() string {
	return r.Type
}

type SshKeypairValue struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

type SshKeypairResponse struct {
	Name  string          `json:"name"`
	Id    string          `json:"id"`
	Value SshKeypairValue `json:"value"`
}

func (res SshKeypairResponse) JsonString() string {
	structBytes, err := json.Marshal(res)
	if err != nil {
		return "{}"
	}
	return fmt.Sprintf("%s", structBytes)
}
