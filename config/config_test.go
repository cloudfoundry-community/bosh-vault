package config_test

import (
	"github.com/zipcar/bosh-vault/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
	"path/filepath"
)

var _ = Describe("Config", func() {
	Describe("Configuration Validation", func() {
		Context("a full valid config", func() {
			var (
				jsonConfigPath string
				yamlConfigPath string
			)
			BeforeEach(func() {
				workingDirectory, _ := os.Getwd()
				jsonConfigPath = filepath.Join(workingDirectory, "configfakes/listen-log-config.json")
				yamlConfigPath = filepath.Join(workingDirectory, "configfakes/listen-log-config.yml")
			})
			It("can read a JSON config correctly", func() {
				bvConfig := config.ParseConfig(&jsonConfigPath)
				Expect(bvConfig.Api.Address).To(Equal("localhost:8000"))
				Expect(bvConfig.Log.Level).To(Equal("ERROR"))
				Expect(bvConfig.Api.DrainTimeout).To(Equal(10))
			})
			It("can read a YML config correctly", func() {
				bvConfig := config.ParseConfig(&yamlConfigPath)
				Expect(bvConfig.Api.Address).To(Equal("localhost:8001"))
				Expect(bvConfig.Log.Level).To(Equal("ERROR"))
				Expect(bvConfig.Api.DrainTimeout).To(Equal(10))
			})
		})
		Context("a partial config with only listen address specified", func() {
			var (
				jsonConfigPath string
				yamlConfigPath string
			)
			BeforeEach(func() {
				workingDirectory, _ := os.Getwd()
				jsonConfigPath = filepath.Join(workingDirectory, "configfakes/listen-no-log-config.json")
				yamlConfigPath = filepath.Join(workingDirectory, "configfakes/listen-no-log-config.yml")
			})
			It("can read a JSON config correctly", func() {
				bvConfig := config.ParseConfig(&jsonConfigPath)
				Expect(bvConfig.Api.Address).To(Equal("localhost:2000"))
				Expect(bvConfig.Log.Level).To(Equal(config.DefaultLogLevel))
			})
			It("can read a YML config correctly", func() {
				bvConfig := config.ParseConfig(&yamlConfigPath)
				Expect(bvConfig.Api.Address).To(Equal("localhost:2001"))
				Expect(bvConfig.Log.Level).To(Equal(config.DefaultLogLevel))
			})
		})
		Context("a partial config with only log level specified", func() {
			var (
				jsonConfigPath string
				yamlConfigPath string
			)
			BeforeEach(func() {
				workingDirectory, _ := os.Getwd()
				jsonConfigPath = filepath.Join(workingDirectory, "configfakes/log-no-listen-config.json")
				yamlConfigPath = filepath.Join(workingDirectory, "configfakes/log-no-listen-config.yml")
			})
			It("can read a JSON config correctly", func() {
				bvConfig := config.ParseConfig(&jsonConfigPath)
				Expect(bvConfig.Api.Address).To(Equal(config.DefaultApiListenAddress))
				Expect(bvConfig.Log.Level).To(Equal("ERROR"))
			})
			It("can read a YML config correctly", func() {
				bvConfig := config.ParseConfig(&yamlConfigPath)
				Expect(bvConfig.Api.Address).To(Equal(config.DefaultApiListenAddress))
				Expect(bvConfig.Log.Level).To(Equal("ERROR"))
			})
		})
		Context("no config file specified", func() {
			It("correctly returns defaults", func() {
				bvConfig := config.ParseConfig(nil)
				Expect(bvConfig.Api.Address).To(Equal(config.DefaultApiListenAddress))
				Expect(bvConfig.Log.Level).To(Equal(config.DefaultLogLevel))
			})
		})
		Context("a non-existent file is specified", func() {
			var (
				fakeConfigPath string
			)
			BeforeEach(func() {
				fakeConfigPath = "/waka/waka/waka.json"
			})
			It("correctly returns defaults", func() {
				bvConfig := config.ParseConfig(&fakeConfigPath)
				Expect(bvConfig.Api.Address).To(Equal(config.DefaultApiListenAddress))
				Expect(bvConfig.Log.Level).To(Equal(config.DefaultLogLevel))
			})
		})
	})

})
