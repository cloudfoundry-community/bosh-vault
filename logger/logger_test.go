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
				bvConfig              config.Configuration
				errorLevelString      string
				errorLevelStringCaps  string
				errorLevelStringMixed string
			)
			BeforeEach(func() {
				bvConfig := config.GetConfig(nil)
				InitializeLogger(bvConfig)
				errorLevelString = "error"
				errorLevelStringCaps = "ERROR"
				errorLevelStringMixed = "ErRoR"
			})

			It("correctly interprets the log level string error", func() {
				bvConfig.LogLevel = errorLevelString
				InitializeLogger(bvConfig)
				Expect(Log.Level).To(Equal(logrus.ErrorLevel))
				Expect(Log.Level).ToNot(Equal(logrus.DebugLevel))
			})

			It("correctly interprets the log level string ERROR", func() {
				bvConfig.LogLevel = errorLevelStringCaps
				InitializeLogger(bvConfig)
				Expect(Log.Level).To(Equal(logrus.ErrorLevel))
				Expect(Log.Level).ToNot(Equal(logrus.DebugLevel))

			})

			It("correctly interprets the log level string ErRoR", func() {
				bvConfig.LogLevel = errorLevelStringMixed
				InitializeLogger(bvConfig)
				Expect(Log.Level).To(Equal(logrus.ErrorLevel))
				Expect(Log.Level).ToNot(Equal(logrus.DebugLevel))
			})
		})
		Context("invalid log level configuration", func() {
			var (
				bvConfig        config.Configuration
				invalidLogLevel string
			)
			BeforeEach(func() {
				bvConfig := config.GetConfig(nil)
				InitializeLogger(bvConfig)
				invalidLogLevel = "waka"
			})
			It("rejects invalid logging level and defaults to DEBUG, since obviously this user needs help", func() {
				bvConfig.LogLevel = invalidLogLevel
				InitializeLogger(bvConfig)
				Expect(Log.Level).To(Equal(logrus.DebugLevel))
			})
		})
	})
})
