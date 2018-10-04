package types

const RsaKeypairType = "rsa"

type RsaKeypairRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func (r *RsaKeypairRequest) IsRsaKeypairRequest() bool {
	return r.Type == RsaKeypairType
}

func (r *RsaKeypairRequest) Generate() error {
	return nil
}
