package types

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/mitchellh/mapstructure"
	"github.com/zipcar/vault-cfcs/logger"
	"github.com/zipcar/vault-cfcs/vault"
	"golang.org/x/crypto/ssh"
)

const SshKeypairType = "ssh"

type SshKeypairRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func SshMarshalVaultData(value SshKeypairValue) map[string]interface{} {
	return map[string]interface{}{
		"public_key":             value.PublicKey,
		"private_key":            value.PrivateKey,
		"public_key_fingerprint": value.PublicKeyFingerprint,
		"type":                   SshKeypairType,
	}
}

func SshUnmarshalVaultData(rawVaultData *vault.SecretResponse) *vault.SecretResponse {
	var keypairResponse SshKeypairValue
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

func (r *SshKeypairRequest) Generate() (GenericCredentialResponse, error) {
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

	sshKeyPair := SshKeypairValue{
		PrivateKey:           string(pemPriv),
		PublicKey:            string(pubKeyBytes),
		PublicKeyFingerprint: ssh.FingerprintLegacyMD5(pubKey),
	}

	secretId, err := vault.StoreSecret(r.Name, SshMarshalVaultData(sshKeyPair))
	resp = SshKeypairResponse{
		Name:  r.Name,
		Id:    secretId,
		Value: sshKeyPair,
	}

	return resp, nil
}

func (r *SshKeypairRequest) CredentialType() string {
	return r.Type
}

type SshKeypairValue struct {
	PublicKey            string `json:"public_key" mapstructure:"public_key"`
	PrivateKey           string `json:"private_key" mapstructure:"private_key"`
	PublicKeyFingerprint string `json:"public_key_fingerprint" mapstructure:"public_key_fingerprint"`
}

type SshKeypairResponse struct {
	Name  string          `json:"name"`
	Id    string          `json:"id"`
	Value SshKeypairValue `json:"value"`
}
