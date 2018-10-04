package types

type PasswordPostReuqest struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type PasswordPostResponse struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}
