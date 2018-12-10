package vault

import (
	"errors"
	"fmt"
	"github.com/hashicorp/vault/api"
	"io"
)

// Currently this method is implemented in Vault's command package kv_helpers.go file but not exported
// there is no "Logical" read method that supports versioned secrets in Vault yet, only a write method.
// There is however, an API endpoint which this helper communicated with
// todo: use a "native" Vault library method directly if one becomes available instead of this borrowed one and delete this file
func kvReadRequest(client *api.Client, path string, params map[string]string) (*api.Secret, error) {
	r := client.NewRequest("GET", "/v1/"+path)
	for k, v := range params {
		r.Params.Set(k, v)
	}
	resp, err := client.RawRequest(r)
	if resp != nil {
		defer resp.Body.Close()
	}
	if resp != nil && resp.StatusCode == 404 {
		secret, parseErr := api.ParseSecret(resp.Body)
		switch parseErr {
		case nil:
		case io.EOF:
			return nil, nil
		default:
			return nil, err
		}
		if secret != nil && (len(secret.Warnings) > 0 || len(secret.Data) > 0) {
			return secret, nil
		}
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return api.ParseSecret(resp.Body)
}

func kvGetMetadata(client *api.Client, name string) (*api.Secret, error) {
	metadataPath := parseMetaDataPath(name)
	metadata, err := client.Logical().Read(metadataPath)
	if err != nil {
		return nil, err
	}

	if metadata == nil {
		return nil, errors.New(fmt.Sprintf("no metadata available for %s", name))
	}

	return metadata, nil
}
