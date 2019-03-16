package types_test

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/zipcar/bosh-vault/store"
	"github.com/zipcar/bosh-vault/types"
	fakes "github.com/zipcar/bosh-vault/types/typesfakes"
	"strings"
	"time"
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
			It("should generate a CA with defaults", func() {
				cert, err := certificateRequest.Generate(&store.SimpleStore{})
				Expect(err).ToNot(HaveOccurred())
				Expect(cert.(types.CertificateRecord).Certificate).To(Equal(cert.(types.CertificateRecord).Ca))
				caBlock, _ := pem.Decode([]byte(cert.(types.CertificateRecord).Ca))
				certificate, err := x509.ParseCertificate(caBlock.Bytes)
				Expect(err).ToNot(HaveOccurred())
				Expect(certificate.IsCA).To(BeTrue())
				Expect(certificate.Subject.CommonName).To(Equal("bosh.io"))
				Expect(strings.Join(certificate.Subject.Country, " ")).To(ContainSubstring(types.CertificateDefaultCountry))
				Expect(strings.Join(certificate.Subject.Organization, " ")).To(ContainSubstring(types.CertificateDefaultOrg))
				Expect(certificate.NotBefore.Add(types.CertificateDefaultTtl * 24 * time.Hour)).To(Equal(certificate.NotAfter))
				Expect(certificate.KeyUsage).To(Equal(x509.KeyUsageCertSign | x509.KeyUsageCRLSign))
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
			It("should generate an intermediate certificate", func() {
				var caName = "datCa"
				caReq := types.CertificateRequest{
					Name: caName,
					Type: types.CertificateType,
					Parameters: types.CertificateParams{
						IsCa:       true,
						CommonName: "goinggoingbackbacktocaca",
					},
				}
				caCert, err := caReq.Generate(&healthySimpleStore)
				Expect(err).ToNot(HaveOccurred())
				_, err = caCert.Store(&healthySimpleStore, caReq.Name)
				Expect(err).ToNot(HaveOccurred())
				intermediate := types.CertificateRequest{
					Name: "intermediate",
					Type: types.CertificateType,
					Parameters: types.CertificateParams{
						IsCa:       true,
						Ca:         caName,
						CommonName: "Tahoe",
					},
				}
				generatedIntCa, err := intermediate.Generate(&healthySimpleStore)
				Expect(err).ToNot(HaveOccurred())
				intCaRecord := generatedIntCa.(types.CertificateRecord)
				Expect(intCaRecord.Ca).To(Equal(caCert.(types.CertificateRecord).Certificate))
				caBlock, _ := pem.Decode([]byte(intCaRecord.Certificate))
				intermediateCertificate, err := x509.ParseCertificate(caBlock.Bytes)
				Expect(intermediateCertificate.IsCA).To(BeTrue())
				Expect(intermediateCertificate.KeyUsage).To(Equal(x509.KeyUsageCertSign | x509.KeyUsageCRLSign))
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
			It("should generate a cert", func() {
				var caName = "datCa"
				caReq := types.CertificateRequest{
					Name: caName,
					Type: types.CertificateType,
					Parameters: types.CertificateParams{
						IsCa:       true,
						CommonName: "goinggoingbackbacktocaca",
					},
				}
				caCert, err := caReq.Generate(&healthySimpleStore)
				Expect(err).ToNot(HaveOccurred())
				_, err = caCert.Store(&healthySimpleStore, caReq.Name)
				Expect(err).ToNot(HaveOccurred())
				leafCertRequest := types.CertificateRequest{
					Name: "leaf",
					Type: types.CertificateType,
					Parameters: types.CertificateParams{
						Ca:         caName,
						CommonName: "Leafy",
					},
				}
				generatedLeafCert, err := leafCertRequest.Generate(&healthySimpleStore)
				Expect(err).ToNot(HaveOccurred())
				leafRecord := generatedLeafCert.(types.CertificateRecord)
				Expect(leafRecord.Ca).To(Equal(caCert.(types.CertificateRecord).Certificate))
				caBlock, _ := pem.Decode([]byte(leafRecord.Certificate))
				leafCert, err := x509.ParseCertificate(caBlock.Bytes)
				Expect(leafCert.IsCA).To(BeFalse())
				Expect(leafCert.KeyUsage).To(Equal(x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature))
			})
		})
	})

})
