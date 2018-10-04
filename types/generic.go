package types

import (
	"encoding/json"
	"errors"
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

func ParseGenericCredentialRequest(requestBody string) (GenericCredentialRequest, error) {
	var g GenericCredentialPostRequest
	err := json.Unmarshal([]byte(requestBody), &g)
	if err != nil {
		fmt.Println("Error unmarshalling!")
	}
	switch g.Type {
	case CertificateType:
		var certificate CertificateRequest
		err = json.Unmarshal([]byte(requestBody), &certificate)
		return &certificate, err
	case PasswordType:
		var password PasswordRequest
		err = json.Unmarshal([]byte(requestBody), &password)
		return &password, err
	case SshKeypairType:
		var ssh SshKeypairRequest
		err = json.Unmarshal([]byte(requestBody), &ssh)
		return &ssh, err
	case RsaKeypairType:
		var rsa RsaKeypairRequest
		err = json.Unmarshal([]byte(requestBody), &rsa)
		return &rsa, err
	default:
		return nil, errors.New(fmt.Sprintf("Request type not supported! Must be one of: %s, %s, %s, %s", CertificateType, PasswordType, SshKeypairType, RsaKeypairType))
	}
}
