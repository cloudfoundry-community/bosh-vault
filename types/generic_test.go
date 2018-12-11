package types_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	bvTypes "github.com/zipcar/bosh-vault/types"
	fakes "github.com/zipcar/bosh-vault/types/typesfakes"
)

var _ = Describe("Generic", func() {
	Describe("ParseCredentialGenerationRequest", func() {
		Context("valid certificate requests", func() {
			It("parses a certificate request object", func() {
				credential, err := bvTypes.ParseCredentialGenerationRequest([]byte(fakes.RegularCertRequestBody))
				Expect(err).ToNot(HaveOccurred())
				Expect(credential).To(BeAssignableToTypeOf((*bvTypes.CertificateRequest)(nil)))

				credential, err = bvTypes.ParseCredentialGenerationRequest([]byte(fakes.IntermediateCaRequestBody))
				Expect(err).ToNot(HaveOccurred())
				Expect(credential).To(BeAssignableToTypeOf((*bvTypes.CertificateRequest)(nil)))

				credential, err = bvTypes.ParseCredentialGenerationRequest([]byte(fakes.RootCaRequestBody))
				Expect(err).ToNot(HaveOccurred())
				Expect(credential).To(BeAssignableToTypeOf((*bvTypes.CertificateRequest)(nil)))
			})
		})

		Context("valid password requests", func() {
			It("parses a password request object", func() {
				credential, err := bvTypes.ParseCredentialGenerationRequest([]byte(fakes.PasswordPostRequestBody))
				Expect(err).ToNot(HaveOccurred())
				Expect(credential).To(BeAssignableToTypeOf((*bvTypes.PasswordRequest)(nil)))
			})
		})

		Context("valid ssh key requests", func() {
			It("parses a ssh keypair object", func() {
				credential, err := bvTypes.ParseCredentialGenerationRequest([]byte(fakes.SshKeypairRequestBody))
				Expect(err).ToNot(HaveOccurred())
				Expect(credential).To(BeAssignableToTypeOf((*bvTypes.SshKeypairRequest)(nil)))
			})
		})

		Context("valid rsa key requests", func() {
			It("parses a rsa keypair object", func() {
				credential, err := bvTypes.ParseCredentialGenerationRequest([]byte(fakes.RsaKeyPairRequestBody))
				Expect(err).ToNot(HaveOccurred())
				Expect(credential).To(BeAssignableToTypeOf((*bvTypes.RsaKeypairRequest)(nil)))
			})
		})
	})
})
