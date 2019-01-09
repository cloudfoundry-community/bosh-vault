package types_test

import (
	"encoding/json"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/zipcar/bosh-vault/types"
	fakes "github.com/zipcar/bosh-vault/types/typesfakes"
)

var _ = Describe("Certificates", func() {

	Describe("Certificate Request Methods", func() {
		Context("A valid CA request", func() {
			var (
				certificateRequest types.CertificateRequest
			)

			JustBeforeEach(func() {
				err := json.Unmarshal([]byte(fakes.RootCaRequestBody), &certificateRequest)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return true when IsRootCaRequest is called", func() {
				Expect(certificateRequest.IsRootCaRequest()).To(BeTrue())
			})
			It("should return false when IsIntermediateCaRequest is called", func() {
				Expect(certificateRequest.IsIntermediateCaRequest()).To(BeFalse())
			})
			It("should return false when IsRegularCertificateRequest is called", func() {
				Expect(certificateRequest.IsRegularCertificateRequest()).To(BeFalse())
			})
		})

		Context("A valid intermediate CA request", func() {
			var (
				certificateRequest types.CertificateRequest
			)

			JustBeforeEach(func() {
				err := json.Unmarshal([]byte(fakes.IntermediateCaRequestBody), &certificateRequest)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return false when IsRootCaRequest is called", func() {
				Expect(certificateRequest.IsRootCaRequest()).To(BeFalse())
			})
			It("should return true when IsIntermediateCaRequest is called", func() {
				Expect(certificateRequest.IsIntermediateCaRequest()).To(BeTrue())
			})
			It("should return false when IsRegularCertRequest is called", func() {
				Expect(certificateRequest.IsRegularCertificateRequest()).To(BeFalse())
			})
		})

		Context("A valid regular cert request", func() {
			var (
				certificateRequest types.CertificateRequest
			)

			JustBeforeEach(func() {
				err := json.Unmarshal([]byte(fakes.RegularCertRequestBody), &certificateRequest)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return false when IsRootCaRequest is called", func() {
				Expect(certificateRequest.IsRootCaRequest()).To(BeFalse())
			})
			It("should return false when IsIntermediateCaRequest is called", func() {
				Expect(certificateRequest.IsIntermediateCaRequest()).To(BeFalse())
			})
			It("should return true when IsRegularCertRequest is called", func() {
				Expect(certificateRequest.IsRegularCertificateRequest()).To(BeTrue())
			})
		})
	})

})
