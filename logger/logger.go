package logger

import (
	"github.com/sirupsen/logrus"
	"github.com/zipcar/bosh-vault/config"
	"os"
)

var Log *logrus.Logger

func InitializeLogger(bvConfig config.Configuration) {
	Log = logrus.New()
	Log.SetFormatter(&logrus.JSONFormatter{})
	logLevel, err := logrus.ParseLevel(bvConfig.LogLevel)
	if err != nil {
		Log.Errorf("error parsing configured log level %s, defaulting to debug", bvConfig.LogLevel)
		logLevel = logrus.DebugLevel
	}
	Log.SetLevel(logLevel)
	Log.Out = os.Stdout
}
