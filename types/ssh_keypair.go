package types

const SshKeypairType = "ssh"

type SshKeypairRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func (s *SshKeypairRequest) Generate() error {
	return nil
}
