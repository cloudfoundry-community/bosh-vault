package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zipcar/vault-cfcs/vault"
)

type GenericCredentialPostRequest struct {
	Name       string          `json:"name"`
	Type       string          `json:"type"`
	Parameters json.RawMessage `json:"parameters, omitempty"`
}

type GenericCredentialPutRequest struct {
	Name  string          `json:"name"`
	Type  string          `json:"type"`
	Value json.RawMessage `json:"value, omitempty"`
}

type GenericCredentialResponse interface{}

type GenericCredentialRecord interface {
	Store(name string) (GenericCredentialResponse, error)
}

type GenericCredentialRequest interface {
	Generate() (GenericCredentialResponse, error)
	Validate() bool
	CredentialType() string
}

func ParseGenericCredentialPostRequest(requestBody []byte) (GenericCredentialRequest, error) {
	var g GenericCredentialPostRequest
	err := json.Unmarshal([]byte(requestBody), &g)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error unmarshaling json request: %s", err.Error()))
	}
	switch g.Type {
	case CertificateType:
		var certificate CertificateRequest
		err = json.Unmarshal(requestBody, &certificate)
		return &certificate, err
	case PasswordType:
		var password PasswordRequest
		err = json.Unmarshal(requestBody, &password)
		return &password, err
	case SshKeypairType:
		var ssh SshKeypairRequest
		err = json.Unmarshal(requestBody, &ssh)
		return &ssh, err
	case RsaKeypairType:
		var rsa RsaKeypairRequest
		err = json.Unmarshal(requestBody, &rsa)
		return &rsa, err
	default:
		return nil, errors.New(fmt.Sprintf("credential request type: %s not supported! Must be one of: %s, %s, %s, %s", g.Type, CertificateType, PasswordType, SshKeypairType, RsaKeypairType))
	}
}

func ParseSecretResponse(vaultSecretResponse vault.SecretResponse) *vault.SecretResponse {
	var secretResp interface{}

	secretType := vaultSecretResponse.Value.(map[string]interface{})["type"].(string)
	switch secretType {
	case PasswordType:
		secretResp = ParseVaultDataAsPassword(&vaultSecretResponse)
	case RsaKeypairType:
		secretResp = ParseVaultDataAsRsaKeypair(&vaultSecretResponse)
	case SshKeypairType:
		secretResp = ParseVaultDataAsSshKeypair(&vaultSecretResponse)
	case CertificateType:
		secretResp = ParseVaultDataAsCertificateRecord(&vaultSecretResponse)
	}

	return secretResp.(*vault.SecretResponse)
}
