package types

const passwordType = "password"

type PasswordPostRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type PasswordPostResponse struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (r *PasswordPostRequest) IsPasswordRequest() bool {
	return r.Type == passwordType
}
