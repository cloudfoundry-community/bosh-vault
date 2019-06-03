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
	"github.com/cloudfoundry-community/bosh-vault/logger"
	"github.com/cloudfoundry-community/bosh-vault/secret"
	"math/big"
	"net"
	"time"
)

const CertificateType = "certificate"
const CertificateDefaultTtl = 365
const CertificateDefaultOrg = "bosh vault"
const CertificateDefaultCountry = "USA"
const CertificateDefaultRsaKeyBits = 2048

type CertificateRequest struct {
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	Parameters CertificateParams `json:"parameters"`
}

type CertificateParams struct {
	CommonName         string   `json:"common_name"`
	IsCa               bool     `json:"is_ca,omitempty"`
	Ca                 string   `json:"ca,omitempty"`
	AlternativeNames   []string `json:"alternative_names,omitempty"`
	ExtendedKeyUsage   []string `json:"extended_key_usage,omitempty"`
	Organization       string   `json:"organization"`
	OrganizationalUnit string   `json:"organizational_unit"`
	Locality           string   `json:"locality"`
	State              string   `json:"state"`
	Country            string   `json:"country"`
	KeyUsage           []string `json:"key_usage"`
	KeyLength          int      `json:"key_length"`
	Duration           int      `json:"duration"`
	SelfSign           bool     `json:"self_sign"`
}

type CertificateResponse struct {
	Name  string            `json:"name"`
	Id    string            `json:"id"`
	Value CertificateRecord `json:"value"`
}

type CertificateRecord struct {
	Certificate string `json:"certificate"`
	Ca          string `json:"ca"`
	PrivateKey  string `json:"private_key"`
}

func (r *CertificateRequest) CredentialType() string {
	return r.Type
}

func (r *CertificateRequest) CredentialName() string {
	return r.Name
}

func (r *CertificateRequest) Validate() bool {
	return r.IsRootCaRequest() || r.IsIntermediateCaRequest() || r.IsRegularCertificateRequest()
}

func (r CertificateRequest) IsRootCaRequest() bool {
	return r.Type == CertificateType &&
		r.Parameters.IsCa &&
		r.Parameters.CommonName != "" &&
		(r.Parameters.Ca == "" || r.Parameters.SelfSign) &&
		len(r.Parameters.AlternativeNames) == 0
}

func (r CertificateRequest) IsIntermediateCaRequest() bool {
	return r.Type == CertificateType &&
		r.Parameters.IsCa &&
		r.Parameters.CommonName != "" &&
		(r.Parameters.Ca != "" || r.Parameters.SelfSign) &&
		len(r.Parameters.AlternativeNames) == 0
}

func (r CertificateRequest) IsRegularCertificateRequest() bool {
	return r.Type == CertificateType &&
		!r.Parameters.IsCa &&
		(r.Parameters.Ca != "" || r.Parameters.SelfSign) &&
		r.Parameters.CommonName != ""
}

func (record CertificateRecord) Store(secretStore secret.Store, name string) (CredentialResponse, error) {
	resp := CertificateResponse{}
	id, err := secretStore.Set(name, map[string]interface{}{
		"certificate": record.Certificate,
		"ca":          record.Ca,
		"private_key": record.PrivateKey,
	})
	if err != nil {
		return resp, err
	}

	resp = CertificateResponse{
		Name:  name,
		Id:    id,
		Value: record,
	}

	return resp, nil
}

func (r *CertificateRequest) Generate(secretStore secret.Store) (CredentialRecordInterface, error) {
	var rootCaCert *x509.Certificate
	var rootCaKey *rsa.PrivateKey
	var err error

	if r.IsRegularCertificateRequest() || r.IsIntermediateCaRequest() {
		if r.Parameters.SelfSign {
			rootCaKey = nil
			rootCaCert = nil
		} else {
			rootCaCert, rootCaKey, err = getRootCaAndKeyByName(r.Parameters.Ca, secretStore)
			if err != nil {
				return nil, err
			}
		}
	}

	switch {
	case r.IsRegularCertificateRequest():
		return r.GenerateRegularCertificate(rootCaCert, rootCaKey)
	case r.IsIntermediateCaRequest():
		return r.GenerateIntermediateCertificate(rootCaCert, rootCaKey)
	case r.IsRootCaRequest():
		return r.GenerateRootCertificate()
	default:
		return nil, errors.New("unable to generate cert, unknown type, make sure to call Validate on the request before trying to Generate")
	}
}

func getRsaKey(keyLength int) (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, keyLength)
	if err != nil {
		return &rsa.PrivateKey{}, errors.New(fmt.Sprintf("Problem generating RSA keypair for cert: %s", err))
	}
	return privateKey, err
}

func newX509CertAndKey(cr *CertificateRequest) (x509.Certificate, *rsa.PrivateKey, error) {

	if cr.Parameters.KeyLength == 0 {
		cr.Parameters.KeyLength = CertificateDefaultRsaKeyBits
	}

	privateKey, err := getRsaKey(cr.Parameters.KeyLength)
	if err != nil {
		return x509.Certificate{}, privateKey, err
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return x509.Certificate{}, privateKey, errors.New(fmt.Sprintf("error generating cert serial number %s", err))
	}

	now := time.Now()

	if cr.Parameters.Duration == 0 {
		cr.Parameters.Duration = CertificateDefaultTtl
	}

	notAfter := now.Add(time.Duration(cr.Parameters.Duration*24) * time.Hour)

	subjectKeyHash := sha1.New()
	subjectKeyHash.Write(privateKey.N.Bytes())
	subjectKeyId := subjectKeyHash.Sum(nil)

	if cr.Parameters.Organization == "" {
		cr.Parameters.Organization = CertificateDefaultOrg
	}

	if cr.Parameters.Country == "" {
		cr.Parameters.Country = CertificateDefaultCountry
	}

	cert := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Country:            []string{cr.Parameters.Country},
			Organization:       []string{cr.Parameters.Organization},
			OrganizationalUnit: []string{cr.Parameters.OrganizationalUnit},
			Locality:           []string{cr.Parameters.Locality},
			Province:           []string{cr.Parameters.State},
			CommonName:         cr.Parameters.CommonName,
		},
		NotBefore:             now,
		NotAfter:              notAfter,
		BasicConstraintsValid: true,
		IsCA:                  cr.Parameters.IsCa,
		SubjectKeyId:          subjectKeyId,
	}

	if cr.Parameters.SelfSign {
		cert.IsCA = true
		cert.KeyUsage = x509.KeyUsageCertSign | x509.KeyUsageCRLSign
	}

	for _, keyUsage := range cr.Parameters.KeyUsage {
		switch keyUsage {
		case "digital_signature":
			cert.KeyUsage |= x509.KeyUsageDigitalSignature
		case "key_encipherment":
			cert.KeyUsage |= x509.KeyUsageKeyEncipherment
		case "non_repudiation":
			cert.KeyUsage |= x509.KeyUsageContentCommitment
		case "data_encipherment":
			cert.KeyUsage |= x509.KeyUsageDataEncipherment
		case "key_agreement":
			cert.KeyUsage |= x509.KeyUsageKeyAgreement
		case "key_cert_sign":
			if cr.IsIntermediateCaRequest() || cr.IsRootCaRequest() {
				cert.KeyUsage |= x509.KeyUsageCertSign
			}
		case "crl_sign":
			cert.KeyUsage |= x509.KeyUsageCRLSign
		case "encipher_only":
			cert.KeyUsage |= x509.KeyUsageEncipherOnly
		case "decipher_only":
			cert.KeyUsage |= x509.KeyUsageDecipherOnly
		default:
			logger.Log.Errorf("Unsupported extended key usage: %s, ignoring", keyUsage)
		}
	}

	for _, extUsage := range cr.Parameters.ExtendedKeyUsage {
		switch extUsage {
		case "client_auth":
			cert.ExtKeyUsage = append(cert.ExtKeyUsage, x509.ExtKeyUsageClientAuth)
		case "server_auth":
			cert.ExtKeyUsage = append(cert.ExtKeyUsage, x509.ExtKeyUsageServerAuth)
		case "code_signing":
			cert.ExtKeyUsage = append(cert.ExtKeyUsage, x509.ExtKeyUsageCodeSigning)
		case "email_protection":
			cert.ExtKeyUsage = append(cert.ExtKeyUsage, x509.ExtKeyUsageEmailProtection)
		case "timestamping":
			cert.ExtKeyUsage = append(cert.ExtKeyUsage, x509.ExtKeyUsageTimeStamping)
		default:
			logger.Log.Errorf("Unsupported extended key usage: %s, ignoring", extUsage)
		}
	}

	return cert, privateKey, nil
}

func getRootCaAndKeyByName(caName string, store secret.Store) (*x509.Certificate, *rsa.PrivateKey, error) {
	rootCaCert := &x509.Certificate{}
	rootCaKey := &rsa.PrivateKey{}
	rawCaResponse, err := store.GetByName(caName)
	if err != nil {
		return rootCaCert, rootCaKey, err
	}

	caRecord := rawCaResponse[0].Value.(map[string]interface{})

	cpb, _ := pem.Decode([]byte(caRecord["certificate"].(string)))
	rootCaCert, err = x509.ParseCertificate(cpb.Bytes)
	if err != nil {
		return rootCaCert, rootCaKey, err
	}

	caPrivKey, _ := pem.Decode([]byte(caRecord["private_key"].(string)))
	rootCaKey, err = x509.ParsePKCS1PrivateKey(caPrivKey.Bytes)
	if err != nil {
		return rootCaCert, rootCaKey, err
	}

	return rootCaCert, rootCaKey, nil
}

func assembleCertRecord(rawCaCert, rawCert, rawKey []byte) CertificateRecord {
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

	return CertificateRecord{
		Certificate: string(pemCert),
		Ca:          string(pemCa),
		PrivateKey:  string(pemPrivateKey),
	}
}

func (r *CertificateRequest) GenerateRegularCertificate(rootCaCert *x509.Certificate, rootCaKey *rsa.PrivateKey) (CredentialRecordInterface, error) {

	certTemplate, privateKey, err := newX509CertAndKey(r)
	if err != nil {
		return nil, err
	}

	// Default key usage for standard TLS certificate
	if len(r.Parameters.KeyUsage) == 0 {
		certTemplate.KeyUsage = x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature
	}

	// Default key usage for "regular" MTLS cert is "server_auth"
	if len(r.Parameters.ExtendedKeyUsage) == 0 {
		certTemplate.ExtKeyUsage = append(certTemplate.ExtKeyUsage, x509.ExtKeyUsageServerAuth)
	}

	if r.Parameters.SelfSign {
		rootCaCert = &certTemplate
		rootCaKey = privateKey
	}

	certTemplate.AuthorityKeyId = rootCaCert.SubjectKeyId

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
		return nil, errors.New(fmt.Sprintf("problem generating the x509 CA cert: %s", err))
	}

	return assembleCertRecord(rootCaCert.Raw, rawCert, x509.MarshalPKCS1PrivateKey(privateKey)), nil
}

func (r *CertificateRequest) GenerateRootCertificate() (CredentialRecordInterface, error) {

	certTemplate, privateKey, err := newX509CertAndKey(r)
	if err != nil {
		return nil, err
	}

	// Default key usage for root TLS certificate
	if len(r.Parameters.KeyUsage) == 0 {
		certTemplate.KeyUsage = x509.KeyUsageCertSign | x509.KeyUsageCRLSign
	}

	certTemplate.AuthorityKeyId = certTemplate.SubjectKeyId

	rawCert, err := x509.CreateCertificate(rand.Reader, &certTemplate, &certTemplate, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("problem generating the x509 CA cert: %s", err))
	}

	return assembleCertRecord(rawCert, rawCert, x509.MarshalPKCS1PrivateKey(privateKey)), nil
}

func (r *CertificateRequest) GenerateIntermediateCertificate(rootCaCert *x509.Certificate, rootCaKey *rsa.PrivateKey) (CredentialRecordInterface, error) {

	certTemplate, privateKey, err := newX509CertAndKey(r)
	if err != nil {
		return nil, err
	}

	// default key usage for intermediate cert
	if len(r.Parameters.KeyUsage) == 0 {
		certTemplate.KeyUsage = x509.KeyUsageCertSign | x509.KeyUsageCRLSign
	}

	if r.Parameters.SelfSign {
		rootCaCert = &certTemplate
		rootCaKey = privateKey
	}

	certTemplate.AuthorityKeyId = rootCaCert.SubjectKeyId

	rawCert, err := x509.CreateCertificate(rand.Reader, &certTemplate, rootCaCert, &privateKey.PublicKey, rootCaKey)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("problem generating the x509 CA cert: %s", err))
	}

	return assembleCertRecord(rootCaCert.Raw, rawCert, x509.MarshalPKCS1PrivateKey(privateKey)), nil
}
