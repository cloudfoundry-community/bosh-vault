package types_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/zipcar/bosh-vault/types"
	fakes "github.com/zipcar/bosh-vault/types/typesfakes"
)

var _ = Describe("Generic", func() {
	Describe("ParseCredentialGenerationRequest", func() {
		Context("valid certificate requests", func() {
			It("parses a certificate request object", func() {
				credential, err := types.ParseCredentialGenerationRequest([]byte(fakes.RegularCertRequestBody))
				Expect(err).ToNot(HaveOccurred())
				Expect(credential).To(BeAssignableToTypeOf((*types.CertificateRequest)(nil)))

				credential, err = types.ParseCredentialGenerationRequest([]byte(fakes.IntermediateCaRequestBody))
				Expect(err).ToNot(HaveOccurred())
				Expect(credential).To(BeAssignableToTypeOf((*types.CertificateRequest)(nil)))

				credential, err = types.ParseCredentialGenerationRequest([]byte(fakes.RootCaRequestBody))
				Expect(err).ToNot(HaveOccurred())
				Expect(credential).To(BeAssignableToTypeOf((*types.CertificateRequest)(nil)))
			})
		})

		Context("valid password requests", func() {
			It("parses a password request object", func() {
				credential, err := types.ParseCredentialGenerationRequest([]byte(fakes.PasswordPostRequestBody))
				Expect(err).ToNot(HaveOccurred())
				Expect(credential).To(BeAssignableToTypeOf((*types.PasswordRequest)(nil)))
			})
		})

		Context("valid ssh key requests", func() {
			It("parses a ssh keypair object", func() {
				credential, err := types.ParseCredentialGenerationRequest([]byte(fakes.SshKeypairRequestBody))
				Expect(err).ToNot(HaveOccurred())
				Expect(credential).To(BeAssignableToTypeOf((*types.SshKeypairRequest)(nil)))
			})
		})

		Context("valid rsa key requests", func() {
			It("parses a rsa keypair object", func() {
				credential, err := types.ParseCredentialGenerationRequest([]byte(fakes.RsaKeyPairRequestBody))
				Expect(err).ToNot(HaveOccurred())
				Expect(credential).To(BeAssignableToTypeOf((*types.RsaKeypairRequest)(nil)))
			})
		})
	})
})
