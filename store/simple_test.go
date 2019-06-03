package store_test

import (
	"github.com/cloudfoundry-community/bosh-vault/store"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var seedData = []store.VaultTestData{
	store.VaultTestData{
		Path: "some_password",
		Value: map[string]interface{}{
			"value": "$$$$$wakawakawaka$$$$$",
		},
		Id:   "eyJuYW1lIjoic29tZV9wYXNzd29yZCIsInZlcnNpb24iOjF9", // ID expects version 1
		Seed: false,                                              // A test expects to write this in for the first time
	},
	store.VaultTestData{
		Path: "some_value",
		Value: map[string]interface{}{
			"value": "theBestValue",
		},
		Id:   "eyJuYW1lIjoic29tZV92YWx1ZSIsInZlcnNpb24iOjB9", // ID expects version 0 or "latest"
		Seed: true,
	},
}

var _ = Describe("Simple Store", func() {
	Describe("store tests", func() {
		Context("health checking", func() {

			It("true when healthy: unsealed and initialized", func() {
				Expect(healthySimpleStore.Healthy()).To(BeTrue())
			})

			It("false when uninitialized", func() {
				Expect(uninitializedVaultSimpleStore.Healthy()).To(BeFalse())
			})

			It("false when sealed", func() {
				Expect(sealedVaultSimpleStore.Healthy()).To(BeFalse())
			})

		})

		Context("secrets", func() {

			BeforeEach(func() {
				for _, data := range seedData {
					if data.Seed {
						_, _ = healthySimpleStore.Set(data.Path, data.Value)
					}
				}
			})

			AfterEach(func() {
				for _, data := range seedData {
					_ = healthySimpleStore.DeleteByName(data.Path)
				}

			})

			It("can write secrets to Vault", func() {
				for _, data := range seedData {
					if !data.Seed {
						// confirm the secret doesn't ALREADY exist
						exists := healthySimpleStore.Exists(data.Path)
						Expect(exists).To(BeFalse())

						// write the secret
						id, err := healthySimpleStore.Set(data.Path, data.Value)
						Expect(err).ToNot(HaveOccurred())
						Expect(id).ToNot(BeEmpty())
						Expect(id).To(Equal(data.Id))

						// secret should exist
						exists = healthySimpleStore.Exists(data.Path)
						Expect(exists).To(BeTrue())

						// secret value should be set correctly
						secret, err := healthySimpleStore.GetLatestByName(data.Path)
						Expect(err).ToNot(HaveOccurred())
						Expect(secret.Value).To(Equal(data.Value))
					}
				}
			})

			It("can tell when a secret exists and when it doesn't", func() {
				for _, data := range seedData {
					if data.Seed {
						exists := healthySimpleStore.Vault.Exists(data.Path)
						Expect(exists).To(BeTrue())
					} else {
						exists := healthySimpleStore.Vault.Exists(data.Path)
						Expect(exists).To(BeFalse())
					}
				}
			})

			It("returns not found when asked for secrets that don't exist by name", func() {
				_, err := healthySimpleStore.GetLatestByName("/a/totally/bs/path")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not found"))
			})

			It("returns not found when asked for secrets that don't exist by id", func() {
				_, err := healthySimpleStore.GetById("eyJuYW1lIjoiL0RpcmVjdG9yL25naW54L3NvbWVfcGFzc3dvcmQiLCJ2ZXJzaW9uIjoxfQ==")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not found"))
			})

			It("can get secrets by id", func() {
				for _, data := range seedData {
					if data.Seed {
						secret, err := healthySimpleStore.GetById(data.Id)
						Expect(err).ToNot(HaveOccurred())
						Expect(secret.Value).To(Equal(data.Value))
					}
				}
			})

			It("can delete secrets from Vault", func() {
				for _, data := range seedData {
					if data.Seed {
						err := healthySimpleStore.Vault.Delete(data.Path)
						Expect(err).ToNot(HaveOccurred())
						exists := healthySimpleStore.Vault.Exists(data.Path)
						Expect(exists).To(BeFalse())
					}
				}
			})

		})
	})
})
