package config

import (
	"github.com/micro/go-config"
	"github.com/micro/go-config/source/file"
)

const default_api_listen_address = "localhost:1337"
const default_log_level = "DEBUG"

type Configuration struct {
	ApiListenAddress string `json:"listen_addr" yaml:"listen_addr"`
	LogLevel         string `json:"log_level" yaml:"log_level"`
}

func GetConfig(configFilePath *string) Configuration {
	var vcfcsConfig Configuration
	vcfcsConfig.ApiListenAddress = default_api_listen_address
	vcfcsConfig.LogLevel = default_log_level

	if *configFilePath == "" {
		return vcfcsConfig
	} else {
		conf := config.NewConfig()
		conf.Load(file.NewSource(
			file.WithPath(*configFilePath),
		))
		conf.Scan(&vcfcsConfig)
		return vcfcsConfig
	}
}
