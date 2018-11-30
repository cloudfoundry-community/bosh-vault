package types_test

import (
	"encoding/json"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	vcfcsTypes "github.com/zipcar/bosh-vault/types"
	fakes "github.com/zipcar/bosh-vault/types/typesfakes"
)

var _ = Describe("Password", func() {
	Describe("Password request validation", func() {
		Context("a valid password post request", func() {
			var (
				PasswordRequest vcfcsTypes.PasswordRequest
			)

			BeforeEach(func() {
				err := json.Unmarshal([]byte(fakes.PasswordPostRequestBody), &PasswordRequest)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns true when Validate is called", func() {
				Expect(PasswordRequest.Validate()).To(BeTrue())
			})
		})
	})
})
