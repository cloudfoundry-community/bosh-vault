package main

import (
	"flag"
	"fmt"
	"github.com/zipcar/vault-cfcs/config"
	"github.com/zipcar/vault-cfcs/logger"
	"github.com/zipcar/vault-cfcs/server"
	"github.com/zipcar/vault-cfcs/version"
	"os"
)

func main() {
	showVersionAndExit := flag.Bool("version", false, "display version and exit")
	configPath := flag.String("config", "", "path to the configuration file")
	flag.Parse()

	if *showVersionAndExit {
		fmt.Println(fmt.Sprintf("vault-cfcs version: %s", version.Version))
		return
	}

	// If config flag wasn't passed check the environment too, if this is empty too GetConfig can deal with it (use defaults)
	if *configPath == "" {
		configFilePathEnvValue := os.Getenv("VCFCS_CONFIG")
		configPath = &configFilePathEnvValue
	}

	vcfcsConfig := config.GetConfig(configPath)

	logger.InitializeLogger(vcfcsConfig)
	logger.Log.Infof("Hello world. I am vault-cfcs version %s", version.Version)
	server.ListenAndServe(vcfcsConfig)
}
