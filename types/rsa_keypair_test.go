package types_test

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"github.com/cloudfoundry-community/bosh-vault/store"
	"github.com/cloudfoundry-community/bosh-vault/types"
	fakes "github.com/cloudfoundry-community/bosh-vault/types/typesfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

			It("generates a real RSA key", func() {
				rsaReq, err := RsaRequest.Generate(&store.SimpleStore{})
				Expect(err).ToNot(HaveOccurred())
				rsaRecord := rsaReq.(types.RsaKeypairRecord)
				Expect(rsaRecord.PrivateKey).ToNot(BeEmpty())
				Expect(rsaRecord.PublicKey).ToNot(BeEmpty())

				privBlock, _ := pem.Decode([]byte(rsaRecord.PrivateKey))
				rsaPriv, err := x509.ParsePKCS1PrivateKey(privBlock.Bytes)
				Expect(err).ToNot(HaveOccurred())

				pubKeyBytes, err := x509.MarshalPKIXPublicKey(rsaPriv.Public())
				Expect(err).ToNot(HaveOccurred())
				pemPublic := pem.EncodeToMemory(&pem.Block{
					Type:  "PUBLIC KEY",
					Bytes: pubKeyBytes,
				})

				Expect(string(pemPublic)).To(Equal(rsaRecord.PublicKey))
			})
		})
	})
})
