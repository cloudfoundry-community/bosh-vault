package vault_test

import (
	"encoding/json"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Vault Client", func() {
	Describe("api", func() {
		Context("single vault node", func() {
			BeforeEach(func() {
				for _, data := range seedData {
					if data.Seed {
						_, _ = healthyVault.Set(data.Path, data.Value)
					}
				}
			})
			It("Can correctly report its health", func() {
				Expect(healthyVault.Healthy()).To(BeTrue())
			})
			It("Can write data to Vault", func() {
				for _, data := range seedData {
					if !data.Seed {
						vr, err := healthyVault.Set(data.Path, data.Value)
						Expect(err).ToNot(HaveOccurred())
						Expect(vr["version"].(json.Number)).ToNot(BeEmpty())
					}
				}
			})
			It("Can correctly determine if something exists", func() {
				for _, data := range seedData {
					if data.Seed {
						Expect(healthyVault.Exists(data.Path)).To(BeTrue())
					}
				}
			})
			It("Can read meta data from vault", func() {
				_, err := healthyVault.Set(seedForKnownMetaDataTest.Path, seedForKnownMetaDataTest.Value)
				Expect(err).ToNot(HaveOccurred())

				meta, err := healthyVault.GetMetadata(seedForKnownMetaDataTest.Path)
				Expect(err).ToNot(HaveOccurred())
				Expect(meta).ToNot(BeEmpty())
				Expect(meta["current_version"]).To(Equal(json.Number("1")))
				Expect(meta["versions"].(map[string]interface{})["1"].(map[string]interface{})["destroyed"]).To(BeFalse())
			})
			It("Can read data from vault", func() {
				for _, data := range seedData {
					if data.Seed {
						vr, err := healthyVault.Get(data.Path, map[string]string{
							"version": "1",
						})
						Expect(err).ToNot(HaveOccurred())
						Expect(vr["data"]).To(Equal(data.Value))
					}
				}
			})
		})
	})
})
