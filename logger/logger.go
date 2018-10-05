package logger

import (
	"github.com/sirupsen/logrus"
	"github.com/zipcar/vault-cfcs/config"
	"os"
)

var Log *logrus.Logger

func InitializeLogger(vcfcsConfig config.Configuration) {
	Log = logrus.New()
	Log.SetFormatter(&logrus.JSONFormatter{})
	logLevel, err := logrus.ParseLevel(vcfcsConfig.LogLevel)
	if err != nil {
		Log.Error("Error parsing configured log level, defaulting to ERROR.")
		logLevel = logrus.ErrorLevel
	}
	Log.SetLevel(logLevel)
	Log.Out = os.Stdout
}
