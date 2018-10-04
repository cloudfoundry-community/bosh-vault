package server

import (
	"github.com/zipcar/vault-cfcs/logger"
)

func ListenAndServe() {
	logger.InitializeLogger()
	logger.Log.Print("Server starting")
}
