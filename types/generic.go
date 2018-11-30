package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zipcar/bosh-vault/vault"
)

type GenericCredentialGenerationRequest struct {
	Name       string          `json:"name"`
	Type       string          `json:"type"`
	Parameters json.RawMessage `json:"parameters, omitempty"`
}

type GenericCredentialSetRequest struct {
	Name  string          `json:"name"`
	Type  string          `json:"type"`
	Value json.RawMessage `json:"value, omitempty"`
}

type CredentialSetRequest struct {
	Name   string
	Type   string
	Record CredentialRecordInterface
}

type CredentialResponse interface{}

type CredentialRecordInterface interface {
	Store(name string) (CredentialResponse, error)
}

type CredentialGenerationRequest interface {
	Generate() (CredentialResponse, error)
	Validate() bool
	CredentialType() string
}

func ParseCredentialSetRequest(requestBody []byte) (CredentialSetRequest, error) {
	var g GenericCredentialSetRequest

	err := json.Unmarshal(requestBody, &g)
	if err != nil {
		return CredentialSetRequest{}, errors.New(fmt.Sprintf("error unmarshaling json request: %s", err.Error()))
	}

	var record CredentialRecordInterface

	switch g.Type {
	case CertificateType:
		record = &CertificateRecord{}
	case PasswordType:
		// PasswordRecords are just fancy strings and can't be initialized like structs so we gotta do this
		var passRecord PasswordRecord
		record = &passRecord
	case SshKeypairType:
		record = &SshKeypairRecord{}
	case RsaKeypairType:
		record = &RsaKeypairRecord{}
	default:
		return CredentialSetRequest{}, errors.New(fmt.Sprintf("credential set request type: %s not supported! Must be one of: %s, %s, %s, %s", g.Type, CertificateType, PasswordType, SshKeypairType, RsaKeypairType))
	}

	err = json.Unmarshal(g.Value, &record)

	return CredentialSetRequest{
		Name:   g.Name,
		Type:   g.Type,
		Record: record,
	}, err
}

func ParseCredentialGenerationRequest(requestBody []byte) (CredentialGenerationRequest, error) {
	var g GenericCredentialGenerationRequest
	err := json.Unmarshal([]byte(requestBody), &g)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error unmarshaling json request: %s", err.Error()))
	}

	var req CredentialGenerationRequest
	switch g.Type {
	case CertificateType:
		req = &CertificateRequest{}
	case PasswordType:
		req = &PasswordRequest{}
	case SshKeypairType:
		req = &SshKeypairRequest{}
	case RsaKeypairType:
		req = &RsaKeypairRequest{}
	default:
		return nil, errors.New(fmt.Sprintf("credential request type: %s not supported! Must be one of: %s, %s, %s, %s", g.Type, CertificateType, PasswordType, SshKeypairType, RsaKeypairType))
	}

	err = json.Unmarshal(requestBody, &req)
	return req, err
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
