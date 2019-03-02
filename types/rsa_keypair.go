package types

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"github.com/zipcar/bosh-vault/secret"
)

const RsaKeypairType = "rsa"
const RsaKeySizeBits = 2048

type RsaKeypairRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func (record RsaKeypairRecord) Store(secretStore secret.Store, name string) (CredentialResponse, error) {
	var resp RsaKeypairResponse
	secretId, err := secretStore.Set(name, map[string]interface{}{
		"public_key":  record.PublicKey,
		"private_key": record.PrivateKey,
	})

	if err != nil {
		return resp, err
	}

	resp = RsaKeypairResponse{
		Name:  name,
		Id:    secretId,
		Value: record,
	}

	return resp, nil
}

func (r *RsaKeypairRequest) Validate() bool {
	return r.Type == RsaKeypairType
}

func (r *RsaKeypairRequest) Generate(secretStore secret.Store) (CredentialRecordInterface, error) {
	privKey, err := rsa.GenerateKey(rand.Reader, RsaKeySizeBits)
	if err != nil {
		return nil, err
	}

	pemPriv := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privKey),
	})

	pubKeyBytes, err := asn1.Marshal(privKey.PublicKey)
	if err != nil {
		return nil, err
	}
	pemPublic := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	rsaKeyPair := RsaKeypairRecord{
		PublicKey:  string(pemPublic),
		PrivateKey: string(pemPriv),
	}

	return rsaKeyPair, nil
}

func (r *RsaKeypairRequest) CredentialType() string {
	return r.Type
}

func (r *RsaKeypairRequest) CredentialName() string {
	return r.Name
}

type RsaKeypairRecord struct {
	PublicKey  string `json:"public_key" mapstructure:"public_key"`
	PrivateKey string `json:"private_key" mapstructure:"private_key"`
}

type RsaKeypairResponse struct {
	Name  string           `json:"name"`
	Id    string           `json:"id"`
	Value RsaKeypairRecord `json:"value"`
}
