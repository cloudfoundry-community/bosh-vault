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

func RsaMarshalVaultData(value RsaKeypairValue) map[string]interface{} {
	return map[string]interface{}{
		"public_key":  value.PublicKey,
		"private_key": value.PrivateKey,
		"type":        RsaKeypairType,
	}
}

func RsaUnmarshalVaultData(rawVaultData *vault.SecretResponse) *vault.SecretResponse {
	var keypairResponse RsaKeypairValue
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

	rsaKeyPair := RsaKeypairValue{
		PublicKey:  string(pemPublic),
		PrivateKey: string(pemPriv),
	}

	secretId, err := vault.StoreSecret(r.Name, RsaMarshalVaultData(rsaKeyPair))
	if err != nil {
		return resp, err
	}

	resp = RsaKeypairResponse{
		Name:  r.Name,
		Id:    secretId,
		Value: rsaKeyPair,
	}

	return resp, nil
}

func (r *RsaKeypairRequest) CredentialType() string {
	return r.Type
}

type RsaKeypairValue struct {
	PublicKey  string `json:"public_key" mapstructure:"public_key"`
	PrivateKey string `json:"private_key" mapstructure:"private_key"`
}

type RsaKeypairResponse struct {
	Name  string          `json:"name"`
	Id    string          `json:"id"`
	Value RsaKeypairValue `json:"value"`
}
