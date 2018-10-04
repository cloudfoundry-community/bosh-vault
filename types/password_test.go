package types_test

import (
	"encoding/json"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	vcfcsTypes "github.com/zipcar/vault-cfcs/types"
)

var PasswordPostRequestBody = `
{
  "name": "mypasswd",
  "type": "password"
}
`
var _ = Describe("Password", func() {
	var (
		PasswordRequest vcfcsTypes.PasswordPostRequest
	)
	BeforeEach(func() {
		err := json.Unmarshal([]byte(PasswordPostRequestBody), &PasswordRequest)
		if err != nil {
			fmt.Println(err.Error())
		}
	})
	Describe("Password request validation", func() {
		Context("a valid password post request", func() {
			It("returns true when IsPasswordRequest is called", func() {
				Expect(PasswordRequest.IsPasswordRequest()).To(BeTrue())
			})
		})
	})
})
