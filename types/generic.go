package types

import (
	"encoding/json"
	"fmt"
)

type GenericCredentialPostRequest struct {
	Name       string          `json:"name"`
	Type       string          `json:"type"`
	Parameters json.RawMessage `json:"parameters, omitempty"`
}

type GenericCredentialRequest interface {
	Generate() error
}

func ParseGenericCredentialRequest(requestBody string) GenericCredentialRequest {
	var g GenericCredentialPostRequest
	err := json.Unmarshal([]byte(requestBody), &g)
	if err != nil {
		fmt.Println("Error unmarshalling!")
	}
	switch g.Type {
	case CertificateType:
		var certificate CertificateRequest
		json.Unmarshal([]byte(requestBody), &certificate)
		return &certificate
	case PasswordType:
		var password PasswordRequest
		json.Unmarshal([]byte(requestBody), &password)
		return &password
	case SshKeypairType:
		var ssh SshKeypairRequest
		json.Unmarshal([]byte(requestBody), &ssh)
		return &ssh
	default:
		return nil
	}
}
