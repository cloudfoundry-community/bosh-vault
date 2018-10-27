package vault

// The config server API is meant to return an Id that can serve as an immutable reference to a specific version of a
// secret. Vault's KV2 secret engine is able to understand secret versions that exist at a single path without UUIDs.
// Moreover, Vault provides no built in way to reference a secret other than by it's path; which is probably best.
//
// To avoid storing a mapping of generated UUIDs to paths and versions in Vault itself we encode all the information to
// fetch a given secret (none of which is itself secret) as the ID. The config server specification only states that the
// ID be an immutable and reference to a specific secret version, which this encoded ID is. This also allows us to take
// advantage of Vault's built in secret versioning, which is super cool.

import (
	"encoding/base64"
	"encoding/json"
	"github.com/zipcar/vault-cfcs/logger"
)

type VersionedSecretMetaData struct {
	Name    string      `json:"name"`
	Path    string      `json:"path"`
	Version json.Number `json:"version"`
}

func noPaddingBase64Encoding() *base64.Encoding {
	return base64.StdEncoding.WithPadding(base64.NoPadding)
}

func EncodeId(record VersionedSecretMetaData) (string, error) {
	recordBytes, err := json.Marshal(record)
	if err != nil {
		logger.Log.Errorf("problem marshaling versioned secret meta data: %+v", record)
		return "", err
	}
	return noPaddingBase64Encoding().EncodeToString(recordBytes), nil
}

func DecodeId(id string) (VersionedSecretMetaData, error) {
	var record VersionedSecretMetaData
	recordBytes, err := noPaddingBase64Encoding().DecodeString(id)
	if err != nil {
		logger.Log.Errorf("problem decoding id: %s", id)
		return record, err
	}
	err = json.Unmarshal(recordBytes, &record)
	if err != nil {
		logger.Log.Errorf("problem unmarshaling id bytes into versioned secret meta data: %s", id)
	}
	return record, nil
}
