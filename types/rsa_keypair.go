package types

const RsaKeypairType = "rsa"

type RsaKeypairRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func (r *RsaKeypairRequest) Validate() bool {
	return r.Type == RsaKeypairType
}

func (r *RsaKeypairRequest) Generate() error {
	return nil
}

func (r *RsaKeypairRequest) CredentialType() string {
	return r.Type
}
