package logger

import (
	"github.com/sirupsen/logrus"
	"github.com/zipcar/bosh-vault/config"
	"os"
)

var Log *logrus.Logger

func InitializeLogger(vcfcsConfig config.Configuration) {
	Log = logrus.New()
	Log.SetFormatter(&logrus.JSONFormatter{})
	logLevel, err := logrus.ParseLevel(vcfcsConfig.LogLevel)
	if err != nil {
		Log.Errorf("error parsing configured log level %s, defaulting to debug", vcfcsConfig.LogLevel)
		logLevel = logrus.DebugLevel
	}
	Log.SetLevel(logLevel)
	Log.Out = os.Stdout
}
