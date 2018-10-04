package main

import (
	"github.com/zipcar/vault-cfcs/logger"
	"github.com/zipcar/vault-cfcs/server"
)

func main() {
	logger.InitializeLogger()
	logger.Log.Info("Hello world. I am vault-cfcs.")
	server.ListenAndServe()
}
