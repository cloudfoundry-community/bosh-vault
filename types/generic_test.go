package types_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	vcfcsTypes "github.com/zipcar/vault-cfcs/types"
	fakes "github.com/zipcar/vault-cfcs/types/typesfakes"
)

var _ = Describe("Generic", func() {
	Describe("ParseGenericCredentialRequest", func() {
		Context("valid certificate requests", func() {
			It("parses a certificate request object", func() {
				credential, err := vcfcsTypes.ParseGenericCredentialRequest([]byte(fakes.RegularCertRequestBody))
				Expect(err).ToNot(HaveOccurred())
				Expect(credential).To(BeAssignableToTypeOf((*vcfcsTypes.CertificateRequest)(nil)))

				credential, err = vcfcsTypes.ParseGenericCredentialRequest([]byte(fakes.IntermediateCaRequestBody))
				Expect(err).ToNot(HaveOccurred())
				Expect(credential).To(BeAssignableToTypeOf((*vcfcsTypes.CertificateRequest)(nil)))

				credential, err = vcfcsTypes.ParseGenericCredentialRequest([]byte(fakes.RootCaRequestBody))
				Expect(err).ToNot(HaveOccurred())
				Expect(credential).To(BeAssignableToTypeOf((*vcfcsTypes.CertificateRequest)(nil)))
			})
		})

		Context("valid password requests", func() {
			It("parses a password request object", func() {
				credential, err := vcfcsTypes.ParseGenericCredentialRequest([]byte(fakes.PasswordPostRequestBody))
				Expect(err).ToNot(HaveOccurred())
				Expect(credential).To(BeAssignableToTypeOf((*vcfcsTypes.PasswordRequest)(nil)))
			})
		})

		Context("valid ssh key requests", func() {
			It("parses a ssh keypair object", func() {
				credential, err := vcfcsTypes.ParseGenericCredentialRequest([]byte(fakes.SshKeypairRequestBody))
				Expect(err).ToNot(HaveOccurred())
				Expect(credential).To(BeAssignableToTypeOf((*vcfcsTypes.SshKeypairRequest)(nil)))
			})
		})

		Context("valid rsa key requests", func() {
			It("parses a rsa keypair object", func() {
				credential, err := vcfcsTypes.ParseGenericCredentialRequest([]byte(fakes.RsaKeyPairRequestBody))
				Expect(err).ToNot(HaveOccurred())
				Expect(credential).To(BeAssignableToTypeOf((*vcfcsTypes.RsaKeypairRequest)(nil)))
			})
		})
	})
})
