package types_test

import (
	"encoding/json"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	vcfcsTypes "github.com/zipcar/vault-cfcs/types"
	fakes "github.com/zipcar/vault-cfcs/types/typesfakes"
)

var _ = Describe("RSA", func() {
	Describe("RSA keypair request validation", func() {
		Context("a valid rsa post request", func() {
			var (
				RsaRequest vcfcsTypes.RsaKeypairRequest
			)

			BeforeEach(func() {
				err := json.Unmarshal([]byte(fakes.RsaKeyPairRequestBody), &RsaRequest)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns true when IsPasswordRequest is called", func() {
				Expect(RsaRequest.IsRsaKeypairRequest()).To(BeTrue())
			})
		})
	})
})
