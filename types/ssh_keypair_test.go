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
	"golang.org/x/crypto/ssh"
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

			It("generates a real ssh key", func() {
				cred, err := SshRequest.Generate(&store.SimpleStore{})
				Expect(err).ToNot(HaveOccurred())
				Expect(cred.(types.SshKeypairRecord).PublicKey).ToNot(BeEmpty())
				Expect(cred.(types.SshKeypairRecord).PrivateKey).ToNot(BeEmpty())
				Expect(cred.(types.SshKeypairRecord).PublicKeyFingerprint).ToNot(BeEmpty())
				privKeyPem, _ := pem.Decode([]byte(cred.(types.SshKeypairRecord).PrivateKey))
				privKey, err := x509.ParsePKCS1PrivateKey(privKeyPem.Bytes)
				Expect(err).ToNot(HaveOccurred())
				pubKey, err := ssh.NewPublicKey(&privKey.PublicKey)
				pubKeyBytes := ssh.MarshalAuthorizedKey(pubKey)
				Expect(string(pubKeyBytes)).To(Equal(cred.(types.SshKeypairRecord).PublicKey))
				Expect(ssh.FingerprintLegacyMD5(pubKey)).To(Equal(cred.(types.SshKeypairRecord).PublicKeyFingerprint))
			})
		})
	})
})
