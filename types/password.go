package types

import (
	"encoding/json"
	"fmt"
)

const PasswordType = "password"

type PasswordRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func (r *PasswordRequest) Validate() bool {
	return r.Type == PasswordType
}

func (r *PasswordRequest) Generate() (GenericCredentialResponse, error) {
	var respObj PasswordPostResponse
	json.Unmarshal([]byte(fmt.Sprintf(`{
		"id":    "1337",
		"name":  "%s",
		"value": "The most secure password the_world has-ever-known!1!!"
		}`, r.Name)), &respObj)
	return respObj, nil
}

func (r *PasswordRequest) CredentialType() string {
	return r.Type
}

type PasswordPostResponse struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (res PasswordPostResponse) JsonString() string {
	structBytes, err := json.Marshal(res)
	if err != nil {
		return "{}"
	}
	return fmt.Sprintf("%s", structBytes)
}
