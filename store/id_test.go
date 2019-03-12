package store_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/zipcar/bosh-vault/logger"
	"github.com/zipcar/bosh-vault/store"
	"github.com/zipcar/bosh-vault/store/storefakes"
	"io/ioutil"
)

var _ = Describe("Store ID", func() {
	BeforeSuite(func() {
		logger.Log = logrus.New()
		logger.Log.Out = ioutil.Discard
	})

	Describe("Id Encoding", func() {
		Context("valid secret metadata", func() {
			It("properly encodes metadata to id", func() {
				id, err := store.EncodeId(storefakes.ValidSecretMetadata)
				Expect(err).To(BeNil())
				Expect(id).To(Equal(storefakes.ValidSecretMetadataId))
			})
		})
	})
	Describe("Id Decoding", func() {
		Context("valid secret id", func() {
			It("properly decodes id to metadata", func() {
				metadata, err := store.DecodeId(storefakes.ValidSecretMetadataId)
				Expect(err).To(BeNil())
				Expect(metadata).To(Equal(storefakes.ValidSecretMetadata))
			})
		})
		Context("invalid secret id", func() {
			It("returns an error for a non base64 string", func() {
				_, err := store.DecodeId("$$$wakawakawaka$$$")
				Expect(err).To(Not(BeNil()))
			})
		})
	})
})
