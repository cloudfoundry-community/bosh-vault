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

type GenericCredential interface {
	Generate() error
}

func ParseGenericCredentialRequest(requestBody string) GenericCredential {
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
	default:
		return nil
	}
}
