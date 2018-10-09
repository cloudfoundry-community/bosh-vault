package types

import (
	"encoding/json"
	"errors"
	"fmt"
)

const CertificateType = "certificate"

type CertificateRequest struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Parameters struct {
		CommonName       string   `json:"common_name"`
		IsCa             bool     `json:"is_ca, omitempty"`
		Ca               string   `json:"ca, omitempty"`
		AlternativeNames []string `json:"alternative_names, omitempty"`
	}
}

func (r CertificateRequest) IsRootCaRequest() bool {
	return r.Type == CertificateType &&
		r.Parameters.IsCa &&
		r.Parameters.CommonName != "" &&
		r.Parameters.Ca == "" &&
		len(r.Parameters.AlternativeNames) == 0
}

func (r CertificateRequest) IsIntermediateCaRequest() bool {
	return r.Type == CertificateType &&
		r.Parameters.IsCa &&
		r.Parameters.CommonName != "" &&
		r.Parameters.Ca != "" &&
		len(r.Parameters.AlternativeNames) == 0
}

func (r CertificateRequest) IsRegularCertificateRequest() bool {
	return r.Type == CertificateType &&
		!r.Parameters.IsCa &&
		r.Parameters.Ca != "" &&
		r.Parameters.CommonName != ""
}

func (r *CertificateRequest) Generate() (GenericCredentialResponse, error) {
	switch {
	case r.IsRegularCertificateRequest():
		return r.GenerateRegularCertificate()
	case r.IsIntermediateCaRequest():
		return r.GenerateIntermediateCertificate()
	case r.IsRootCaRequest():
		return r.GenerateRootCertificate()
	default:
		return nil, errors.New("unable to generate cert, unknown type, make sure to call Validate on the request before trying to Generate.")
	}
}

func (r *CertificateRequest) GenerateRegularCertificate() (GenericCredentialResponse, error) {
	var resp CertificatePostResponse
	err := json.Unmarshal([]byte(fmt.Sprintf(`{
		"name": "%s",
		"id": "1337",
		"value": {
			"certificate": "a great cert the best",
			"ca": "a great ca the most trustworthy",
			"private_key": "the most secure private key"
		}
}`, r.Name)), &resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (r *CertificateRequest) GenerateRootCertificate() (GenericCredentialResponse, error) {
	var resp CertificatePostResponse
	err := json.Unmarshal([]byte(fmt.Sprintf(`{
		"name": "%s",
		"id": "1337",
		"value": {
			"certificate": "a great cert for a root the best",
			"ca": "a great ca for a root the most trustworthy",
			"private_key": "the most secure private key for a root"
		}
}`, r.Name)), &resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (r *CertificateRequest) GenerateIntermediateCertificate() (GenericCredentialResponse, error) {
	var resp CertificatePostResponse
	err := json.Unmarshal([]byte(fmt.Sprintf(`{
		"name": "%s",
		"id": "1337",
		"value": {
			"certificate": "a great intermediate cert the best",
			"ca": "a great ca for an intermediate the most trustworthy",
			"private_key": "the most secure private key for an intermediate"
		}
}`, r.Name)), &resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (r *CertificateRequest) CredentialType() string {
	return r.Type
}

func (r *CertificateRequest) Validate() bool {
	return r.IsRootCaRequest() || r.IsIntermediateCaRequest() || r.IsRegularCertificateRequest()
}

type CertificatePostResponse struct {
	Name  string `json:"name"`
	Id    string `json:"id"`
	Value struct {
		Certificate string `json:"certificate"`
		Ca          string `json:"ca"`
		PrivateKey  string `json:"private_key"`
	} `json:"value"`
}

func (res CertificatePostResponse) JsonString() string {
	structBytes, err := json.Marshal(res)
	if err != nil {
		return "{}"
	}
	return fmt.Sprintf("%s", structBytes)
}
