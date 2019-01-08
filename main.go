package main

import (
	"flag"
	"fmt"
	"github.com/zipcar/bosh-vault/config"
	"github.com/zipcar/bosh-vault/logger"
	"github.com/zipcar/bosh-vault/server"
	"github.com/zipcar/bosh-vault/store"
	"github.com/zipcar/bosh-vault/version"
	"os"
)

func main() {
	showVersionAndExit := flag.Bool("version", false, "display version and exit")
	configPath := flag.String("config", "", "path to the configuration file")
	flag.Parse()

	if *showVersionAndExit {
		fmt.Println(fmt.Sprintf("bosh-vault version: %s", version.Version))
		return
	}

	// If config flag wasn't passed check the environment too, if this is empty too GetConfig can deal with it (use defaults)
	if *configPath == "" {
		configFilePathEnvValue := os.Getenv("BV_CONFIG")
		configPath = &configFilePathEnvValue
	}

	bvConfig := config.GetConfig(configPath)

	logger.InitializeLogger(bvConfig)
	logger.Log.Infof("Hello world. I am bosh-vault version %s", version.Version)
	logger.Log.Debugf("Config: %+v", bvConfig)

	store.Initialize(bvConfig)

	server.ListenAndServe(bvConfig)
}
