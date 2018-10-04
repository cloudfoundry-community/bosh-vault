package types

const SshKeypairType = "ssh"

type SshKeypairRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func (r *SshKeypairRequest) Validate() bool {
	return r.Type == SshKeypairType
}

func (r *SshKeypairRequest) Generate() error {
	return nil
}

func (r *SshKeypairRequest) CredentialType() string {
	return r.Type
}
