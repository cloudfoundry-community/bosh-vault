package types

import (
	gopass "github.com/sethvargo/go-password/password"
	"github.com/zipcar/bosh-vault/secret"
)

const PasswordType = "password"
const PasswordDefaultLength = 30

type PasswordRequest struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Parameters struct {
		Length         int  `json:"length"`
		ExcludeUpper   bool `json:"exclude_upper"`
		ExcludeLower   bool `json:"exclude_lower"`
		ExcludeNumber  bool `json:"exclude_number"`
		IncludeSpecial bool `json:"include_special"`
	} `json:"parameters"`
}

type PasswordRecord string

func (record PasswordRecord) Store(secretStore secret.Store, name string) (CredentialResponse, error) {
	var respObj PasswordResponse
	id, err := secretStore.Set(name, map[string]interface{}{
		"value": record,
	})

	if err != nil {
		return respObj, err
	}

	respObj = PasswordResponse{
		Name:  name,
		Id:    id,
		Value: string(record),
	}

	return respObj, nil
}

func (r *PasswordRequest) numAlphabets() int {
	alphabets := 4
	if r.Parameters.ExcludeLower {
		alphabets -= 1
	}
	if r.Parameters.ExcludeUpper {
		alphabets -= 1
	}
	if r.Parameters.ExcludeNumber {
		alphabets -= 1
	}
	if !r.Parameters.IncludeSpecial {
		alphabets -= 1
	}
	return alphabets
}

func (r *PasswordRequest) Validate() bool {
	return r.Type == PasswordType && r.numAlphabets() > 0
}

func (r *PasswordRequest) Generate(secretStore secret.Store) (CredentialRecordInterface, error) {

	if r.Parameters.Length == 0 {
		r.Parameters.Length = PasswordDefaultLength
	}

	passwordAlphabet := ""

	if !r.Parameters.ExcludeNumber {
		passwordAlphabet += gopass.Digits
	}

	if !r.Parameters.ExcludeUpper {
		passwordAlphabet += gopass.UpperLetters
	}

	if !r.Parameters.ExcludeLower {
		passwordAlphabet += gopass.LowerLetters
	}

	if r.Parameters.IncludeSpecial {
		passwordAlphabet += gopass.Symbols
	}

	// This isn't REALLY just lower letters but it's better than specifying
	// exactly how many of each character we need, other generator alphabets
	// are string types and will default to the empty string
	generator, err := gopass.NewGenerator(&gopass.GeneratorInput{
		LowerLetters: passwordAlphabet,
	})

	// We overload the generator input to avoid length parsing sadness
	// because we build up the entire alphabet in "lower letters" we
	// can tell generate we want no digits, no symbols, and no uppercase letters
	// We prefer this generation library because it is the one used by the existing
	// password generation Vault plugin and maintained by people at Hashicorp
	passValue, err := generator.Generate(r.Parameters.Length, 0, 0, true, true)
	if err != nil {
		return nil, err
	}

	return PasswordRecord(passValue), nil
}

func (r *PasswordRequest) CredentialType() string {
	return r.Type
}

func (r *PasswordRequest) CredentialName() string {
	return r.Name
}

type PasswordResponse struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}
