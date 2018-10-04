package logger

import (
	"github.com/sirupsen/logrus"
	"os"
)

var Log *logrus.Logger

func InitializeLogger() {
	Log = logrus.New()
	Log.Out = os.Stdout
}
