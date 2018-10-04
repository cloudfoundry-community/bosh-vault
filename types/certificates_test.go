package types_test

import (
	"encoding/json"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	vcfcsTypes "github.com/zipcar/vault-cfcs/types"
)

var RootCaRequestBody = `
{
  "name": "my_cert",
  "type": "certificate",
  "parameters": {
    "is_ca": true,
    "common_name": "bosh.io"
  }
}
`

var IntermediateCaRequestBody = `
{
  "name": "my_cert",
  "type": "certificate",
  "parameters": {
    "is_ca": true,
    "ca": "my_ca",
    "common_name": "bosh.io"
  }
}
`

var RegularCertRequestBody = `
{
  "name": "my_cert",
  "type": "certificate",
  "parameters": {
    "ca": "my_ca",
    "common_name": "bosh.io",
    "alternative_names": ["bosh.io", "blah.bosh.io"]
  }
}
`

var _ = Describe("Certificates", func() {
	var (
		RootCaRequest         vcfcsTypes.CertificatePostRequest
		IntermediateCaRequest vcfcsTypes.CertificatePostRequest
		RegularCertRequest    vcfcsTypes.CertificatePostRequest
	)

	BeforeEach(func() {
		err := json.Unmarshal([]byte(RootCaRequestBody), &RootCaRequest)
		if err != nil {
			fmt.Println(err.Error())
		}
		err = json.Unmarshal([]byte(IntermediateCaRequestBody), &IntermediateCaRequest)
		if err != nil {
			fmt.Println(err.Error())
		}
		err = json.Unmarshal([]byte(RegularCertRequestBody), &RegularCertRequest)
		if err != nil {
			fmt.Println(err.Error())
		}
	})

	Describe("Certificate Request Methods", func() {
		Context("A valid CA request", func() {
			It("should return true when IsRootCaRequest is called", func() {
				Expect(RootCaRequest.IsRootCaRequest()).To(BeTrue())
			})
			It("should return false when IsIntermediateCaRequest is called", func() {
				Expect(RootCaRequest.IsIntermediateCaRequest()).To(BeFalse())
			})
			It("should return false when IsRegularCertificateRequest is called", func() {
				Expect(RootCaRequest.IsRegularCertificateRequest()).To(BeFalse())
			})
		})

		Context("A valid intermediate CA request", func() {
			It("should return false when IsRootCaRequest is called", func() {
				Expect(IntermediateCaRequest.IsRootCaRequest()).To(BeFalse())
			})
			It("should return true when IsIntermediateCaRequest is called", func() {
				Expect(IntermediateCaRequest.IsIntermediateCaRequest()).To(BeTrue())
			})
			It("should return false when IsRegularCertRequest is called", func() {
				Expect(IntermediateCaRequest.IsRegularCertificateRequest()).To(BeFalse())
			})
		})

		Context("A valid regular cert request", func() {
			It("should return false when IsRootCaRequest is called", func() {
				Expect(RegularCertRequest.IsRootCaRequest()).To(BeFalse())
			})
			It("should return false when IsIntermediateCaRequest is called", func() {
				Expect(RegularCertRequest.IsIntermediateCaRequest()).To(BeFalse())
			})
			It("should return true when IsRegularCertRequest is called", func() {
				Expect(RegularCertRequest.IsRegularCertificateRequest()).To(BeTrue())
			})
		})
	})

})
