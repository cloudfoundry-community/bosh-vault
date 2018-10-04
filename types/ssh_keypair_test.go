package types_test

import (
	"encoding/json"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	vcfcsTypes "github.com/zipcar/vault-cfcs/types"
	fakes "github.com/zipcar/vault-cfcs/types/typesfakes"
)

var _ = Describe("SSH", func() {
	Describe("SSH keypair request validation", func() {
		Context("a valid ssh post request", func() {
			var (
				SshRequest vcfcsTypes.SshKeypairRequest
			)

			BeforeEach(func() {
				err := json.Unmarshal([]byte(fakes.SshKeypairRequestBody), &SshRequest)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns true when IsPasswordRequest is called", func() {
				Expect(SshRequest.IsSshKeypairRequest()).To(BeTrue())
			})
		})
	})
})
