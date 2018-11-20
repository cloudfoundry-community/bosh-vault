package types

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"github.com/mitchellh/mapstructure"
	"github.com/zipcar/vault-cfcs/logger"
	"github.com/zipcar/vault-cfcs/vault"
)

const RsaKeypairType = "rsa"
const RsaKeySizeBits = 2048

type RsaKeypairRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func (record RsaKeypairRecord) Store(name string) (GenericCredentialResponse, error) {
	var resp RsaKeypairResponse
	secretId, err := vault.StoreSecret(name, map[string]interface{}{
		"public_key":  record.PublicKey,
		"private_key": record.PrivateKey,
		"type":        RsaKeypairType,
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

func ParseVaultDataAsRsaKeypair(rawVaultData *vault.SecretResponse) *vault.SecretResponse {
	var keypairResponse RsaKeypairRecord
	err := mapstructure.Decode(rawVaultData.Value, &keypairResponse)
	if err != nil {
		logger.Log.Error(err)
	}
	rawVaultData.Value = keypairResponse
	return rawVaultData
}

func (r *RsaKeypairRequest) Validate() bool {
	return r.Type == RsaKeypairType
}

func (r *RsaKeypairRequest) Generate() (GenericCredentialResponse, error) {
	var resp RsaKeypairResponse

	privKey, err := rsa.GenerateKey(rand.Reader, RsaKeySizeBits)
	if err != nil {
		return resp, err
	}

	pemPriv := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privKey),
	})

	pubKeyBytes, err := asn1.Marshal(privKey.PublicKey)
	if err != nil {
		return resp, err
	}
	pemPublic := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	rsaKeyPair := RsaKeypairRecord{
		PublicKey:  string(pemPublic),
		PrivateKey: string(pemPriv),
	}

	return rsaKeyPair.Store(r.Name)
}

func (r *RsaKeypairRequest) CredentialType() string {
	return r.Type
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
