package logger_test

import (
	. "github.com/zipcar/bosh-vault/logger"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/zipcar/bosh-vault/config"
)

var _ = Describe("Logger", func() {
	Describe("logging configuration", func() {
		Context("valid log level configurations", func() {
			var (
				vcfcsConfig           config.Configuration
				errorLevelString      string
				errorLevelStringCaps  string
				errorLevelStringMixed string
			)
			BeforeEach(func() {
				vcfcsConfig := config.GetConfig(nil)
				InitializeLogger(vcfcsConfig)
				errorLevelString = "error"
				errorLevelStringCaps = "ERROR"
				errorLevelStringMixed = "ErRoR"
			})

			It("correctly interprets the log level string error", func() {
				vcfcsConfig.LogLevel = errorLevelString
				InitializeLogger(vcfcsConfig)
				Expect(Log.Level).To(Equal(logrus.ErrorLevel))
				Expect(Log.Level).ToNot(Equal(logrus.DebugLevel))
			})

			It("correctly interprets the log level string ERROR", func() {
				vcfcsConfig.LogLevel = errorLevelStringCaps
				InitializeLogger(vcfcsConfig)
				Expect(Log.Level).To(Equal(logrus.ErrorLevel))
				Expect(Log.Level).ToNot(Equal(logrus.DebugLevel))

			})

			It("correctly interprets the log level string ErRoR", func() {
				vcfcsConfig.LogLevel = errorLevelStringMixed
				InitializeLogger(vcfcsConfig)
				Expect(Log.Level).To(Equal(logrus.ErrorLevel))
				Expect(Log.Level).ToNot(Equal(logrus.DebugLevel))
			})
		})
		Context("invalid log level configuration", func() {
			var (
				vcfcsConfig     config.Configuration
				invalidLogLevel string
			)
			BeforeEach(func() {
				vcfcsConfig := config.GetConfig(nil)
				InitializeLogger(vcfcsConfig)
				invalidLogLevel = "waka"
			})
			It("rejects invalid logging level and defaults to DEBUG, since obviously this user needs help", func() {
				vcfcsConfig.LogLevel = invalidLogLevel
				InitializeLogger(vcfcsConfig)
				Expect(Log.Level).To(Equal(logrus.DebugLevel))
			})
		})
	})
})
