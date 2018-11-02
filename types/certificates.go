package types

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/zipcar/vault-cfcs/logger"
	"github.com/zipcar/vault-cfcs/vault"
	"math/big"
	"net"
	"time"
)

const CertificateType = "certificate"
const CertificateDefaultTtl = 365 * 24 * time.Hour
const CertificateDefaultOrg = "vault cfcs"
const CertificateDefaultCountry = "USA"
const CertificateDefaultRsaKeyBits = 3072

type CertificateRequest struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Parameters struct {
		CommonName       string   `json:"common_name"`
		IsCa             bool     `json:"is_ca, omitempty"`
		Ca               string   `json:"ca, omitempty"`
		AlternativeNames []string `json:"alternative_names, omitempty"`
		ExtendedKeyUsage []string `json:"extended_key_usage, omitempty"`
	}
}

type CertificateResponse struct {
	Name  string            `json:"name"`
	Id    string            `json:"id"`
	Value CertificateRecord `json:"value"`
}

type CertificateRecord struct {
	Certificate string `json:"certificate" mapstructure:"certificate"`
	Ca          string `json:"ca" mapstructure:"ca"`
	PrivateKey  string `json:"private_key" mapstructure:"private_key"`
}

func (r *CertificateRequest) CredentialType() string {
	return r.Type
}

func (r *CertificateRequest) Validate() bool {
	return r.IsRootCaRequest() || r.IsIntermediateCaRequest() || r.IsRegularCertificateRequest()
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

func CertificateMarshalVaultData(value CertificateRecord) map[string]interface{} {
	return map[string]interface{}{
		"certificate": value.Certificate,
		"ca":          value.Ca,
		"private_key": value.PrivateKey,
		"type":        CertificateType,
	}
}

func CertificateUnmarshalVaultData(rawVaultData *vault.SecretResponse) *vault.SecretResponse {
	var certResponse CertificateRecord
	err := mapstructure.Decode(rawVaultData.Value, &certResponse)
	if err != nil {
		logger.Log.Error(err)
	}
	rawVaultData.Value = certResponse
	return rawVaultData
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

func getRsaKey() (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, CertificateDefaultRsaKeyBits)
	if err != nil {
		return &rsa.PrivateKey{}, errors.New(fmt.Sprintf("Problem generating RSA keypair for cert: %s", err))
	}
	return privateKey, err
}

func newX509CertAndKey(cr *CertificateRequest) (x509.Certificate, *rsa.PrivateKey, error) {

	privateKey, err := getRsaKey()
	if err != nil {
		return x509.Certificate{}, privateKey, err
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return x509.Certificate{}, privateKey, errors.New(fmt.Sprintf("error generating cert serial number %s", err))
	}

	now := time.Now()
	notAfter := now.Add(CertificateDefaultTtl)

	subjectKeyHash := sha1.New()
	subjectKeyHash.Write(privateKey.N.Bytes())
	subjectKeyId := subjectKeyHash.Sum(nil)

	cert := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Country:      []string{CertificateDefaultCountry},
			Organization: []string{CertificateDefaultOrg},
			CommonName:   cr.Parameters.CommonName,
		},
		NotBefore:             now,
		NotAfter:              notAfter,
		BasicConstraintsValid: true,
		IsCA:                  cr.Parameters.IsCa,
		SubjectKeyId:          subjectKeyId,
	}

	return cert, privateKey, nil
}

func getRootCaAndKeyByName(caName string) (*x509.Certificate, *rsa.PrivateKey, error) {
	rootCaCert := &x509.Certificate{}
	rootCaKey := &rsa.PrivateKey{}

	rawCaResponse, err := vault.FetchSecretByName(caName)
	if err != nil {
		return rootCaCert, rootCaKey, err
	}

	caRecord := ParseSecretResponse(rawCaResponse)

	cpb, _ := pem.Decode([]byte(caRecord.Value.(CertificateRecord).Certificate))
	rootCaCert, err = x509.ParseCertificate(cpb.Bytes)
	if err != nil {
		return rootCaCert, rootCaKey, err
	}

	caPrivKey, _ := pem.Decode([]byte(caRecord.Value.(CertificateRecord).PrivateKey))
	rootCaKey, err = x509.ParsePKCS1PrivateKey(caPrivKey.Bytes)
	if err != nil {
		return rootCaCert, rootCaKey, err
	}

	return rootCaCert, rootCaKey, nil
}

func storeAndReturnCert(name string, rawCaCert, rawCert, rawKey []byte) (CertificateResponse, error) {
	resp := CertificateResponse{}

	pemCa := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: rawCaCert,
	})

	pemCert := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: rawCert,
	})

	pemPrivateKey := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: rawKey,
	})

	certValue := CertificateRecord{
		Certificate: string(pemCert),
		Ca:          string(pemCa),
		PrivateKey:  string(pemPrivateKey),
	}

	id, err := vault.StoreSecret(name, CertificateMarshalVaultData(certValue))
	if err != nil {
		return resp, err
	}

	resp = CertificateResponse{
		Name:  name,
		Id:    id,
		Value: certValue,
	}

	return resp, nil
}

func (r *CertificateRequest) GenerateRegularCertificate() (GenericCredentialResponse, error) {
	var resp CertificateResponse

	certTemplate, privateKey, err := newX509CertAndKey(r)
	if err != nil {
		return resp, err
	}

	rootCaCert, rootCaKey, err := getRootCaAndKeyByName(r.Parameters.Ca)
	if err != nil {
		return resp, err
	}

	certTemplate.KeyUsage = x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature
	certTemplate.AuthorityKeyId = rootCaCert.SubjectKeyId

	// Default key usage for "regular" TLS cert is "server_auth"
	if len(r.Parameters.ExtendedKeyUsage) == 0 {
		r.Parameters.ExtendedKeyUsage = append(r.Parameters.ExtendedKeyUsage, "server_auth")
	}

	for _, extUsage := range r.Parameters.ExtendedKeyUsage {
		switch extUsage {
		case "client_auth":
			certTemplate.ExtKeyUsage = append(certTemplate.ExtKeyUsage, x509.ExtKeyUsageClientAuth)
		case "server_auth":
			certTemplate.ExtKeyUsage = append(certTemplate.ExtKeyUsage, x509.ExtKeyUsageServerAuth)
		default:
			logger.Log.Errorf("Unsupported extended key usage: %s, ignoring", extUsage)
		}
	}

	for _, altName := range r.Parameters.AlternativeNames {
		altNameIp := net.ParseIP(altName)
		if altNameIp == nil {
			certTemplate.DNSNames = append(certTemplate.DNSNames, altName)
		} else {
			certTemplate.IPAddresses = append(certTemplate.IPAddresses, altNameIp)
		}
	}

	rawCert, err := x509.CreateCertificate(rand.Reader, &certTemplate, rootCaCert, &privateKey.PublicKey, rootCaKey)
	if err != nil {
		return resp, errors.New(fmt.Sprintf("problem generating the x509 CA cert: %s", err))
	}

	return storeAndReturnCert(r.Name, rootCaCert.Raw, rawCert, x509.MarshalPKCS1PrivateKey(privateKey))
}

func (r *CertificateRequest) GenerateRootCertificate() (GenericCredentialResponse, error) {
	var resp CertificateResponse

	certTemplate, privateKey, err := newX509CertAndKey(r)
	if err != nil {
		return resp, err
	}

	certTemplate.KeyUsage = x509.KeyUsageCertSign | x509.KeyUsageCRLSign

	certTemplate.AuthorityKeyId = certTemplate.SubjectKeyId

	rawCert, err := x509.CreateCertificate(rand.Reader, &certTemplate, &certTemplate, &privateKey.PublicKey, privateKey)
	if err != nil {
		return resp, errors.New(fmt.Sprintf("problem generating the x509 CA cert: %s", err))
	}

	return storeAndReturnCert(r.Name, rawCert, rawCert, x509.MarshalPKCS1PrivateKey(privateKey))
}

func (r *CertificateRequest) GenerateIntermediateCertificate() (GenericCredentialResponse, error) {
	var resp CertificateResponse

	certTemplate, privateKey, err := newX509CertAndKey(r)
	if err != nil {
		return resp, err
	}

	rootCaCert, rootCaKey, err := getRootCaAndKeyByName(r.Parameters.Ca)
	if err != nil {
		return resp, err
	}

	certTemplate.KeyUsage = x509.KeyUsageCertSign | x509.KeyUsageCRLSign
	certTemplate.AuthorityKeyId = rootCaCert.SubjectKeyId

	rawCert, err := x509.CreateCertificate(rand.Reader, &certTemplate, rootCaCert, &privateKey.PublicKey, rootCaKey)
	if err != nil {
		return resp, errors.New(fmt.Sprintf("problem generating the x509 CA cert: %s", err))
	}

	return storeAndReturnCert(r.Name, rootCaCert.Raw, rawCert, x509.MarshalPKCS1PrivateKey(privateKey))
}
