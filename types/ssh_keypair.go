package types

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/mitchellh/mapstructure"
	"github.com/zipcar/bosh-vault/logger"
	"github.com/zipcar/bosh-vault/store"
	"golang.org/x/crypto/ssh"
)

const SshKeypairType = "ssh"

type SshKeypairRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func (record SshKeypairRecord) ToVaultDataInterface() map[string]interface{} {
	return map[string]interface{}{
		"public_key":             record.PublicKey,
		"private_key":            record.PrivateKey,
		"public_key_fingerprint": record.PublicKeyFingerprint,
		"type":                   SshKeypairType,
	}
}

func (record SshKeypairRecord) Store(name string) (CredentialResponse, error) {
	var resp SshKeypairResponse
	secretId, err := store.SetSecret(name, record.ToVaultDataInterface())
	if err != nil {
		return resp, err
	}

	resp = SshKeypairResponse{
		Name:  name,
		Id:    secretId,
		Value: record,
	}

	return resp, nil
}

func ParseVaultDataAsSshKeypair(rawVaultData *store.SecretResponse) *store.SecretResponse {
	var keypairResponse SshKeypairRecord
	err := mapstructure.Decode(rawVaultData.Value, &keypairResponse)
	if err != nil {
		logger.Log.Error(err)
	}
	rawVaultData.Value = keypairResponse
	return rawVaultData
}

func (r *SshKeypairRequest) Validate() bool {
	return r.Type == SshKeypairType
}

func (r *SshKeypairRequest) Generate() (CredentialResponse, error) {
	var resp SshKeypairResponse

	privKey, err := rsa.GenerateKey(rand.Reader, RsaKeySizeBits)
	if err != nil {
		return resp, err
	}

	pemPriv := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privKey),
	})

	pubKey, err := ssh.NewPublicKey(&privKey.PublicKey)
	if err != nil {
		return nil, err
	}

	pubKeyBytes := ssh.MarshalAuthorizedKey(pubKey)

	sshKeyPair := SshKeypairRecord{
		PrivateKey:           string(pemPriv),
		PublicKey:            string(pubKeyBytes),
		PublicKeyFingerprint: ssh.FingerprintLegacyMD5(pubKey),
	}

	return sshKeyPair.Store(r.Name)
}

func (r *SshKeypairRequest) CredentialType() string {
	return r.Type
}

type SshKeypairRecord struct {
	PublicKey            string `json:"public_key" mapstructure:"public_key"`
	PrivateKey           string `json:"private_key" mapstructure:"private_key"`
	PublicKeyFingerprint string `json:"public_key_fingerprint" mapstructure:"public_key_fingerprint"`
}

type SshKeypairResponse struct {
	Name  string           `json:"name"`
	Id    string           `json:"id"`
	Value SshKeypairRecord `json:"value"`
}
