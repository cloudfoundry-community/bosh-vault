package types_test

import (
	"encoding/json"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/zipcar/bosh-vault/types"
	fakes "github.com/zipcar/bosh-vault/types/typesfakes"
)

var _ = Describe("SSH", func() {
	Describe("SSH keypair request validation", func() {
		Context("a valid ssh post request", func() {
			var (
				SshRequest types.SshKeypairRequest
			)

			BeforeEach(func() {
				err := json.Unmarshal([]byte(fakes.SshKeypairRequestBody), &SshRequest)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns true when Validate is called", func() {
				Expect(SshRequest.Validate()).To(BeTrue())
			})
		})
	})
})
