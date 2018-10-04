package types

const PasswordType = "password"

type PasswordRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type PasswordPostResponse struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (r *PasswordRequest) IsPasswordRequest() bool {
	return r.Type == PasswordType
}

func (r *PasswordRequest) Generate() error {
	return nil
}
