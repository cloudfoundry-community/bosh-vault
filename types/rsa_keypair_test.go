package types_test

import (
	"encoding/json"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/zipcar/bosh-vault/types"
	fakes "github.com/zipcar/bosh-vault/types/typesfakes"
)

var _ = Describe("RSA", func() {
	Describe("RSA keypair request validation", func() {
		Context("a valid rsa post request", func() {
			var (
				RsaRequest types.RsaKeypairRequest
			)

			BeforeEach(func() {
				err := json.Unmarshal([]byte(fakes.RsaKeyPairRequestBody), &RsaRequest)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns true when Validate is called", func() {
				Expect(RsaRequest.Validate()).To(BeTrue())
			})
		})
	})
})
