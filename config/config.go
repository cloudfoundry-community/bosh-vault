package config

import (
	"github.com/micro/go-config"
	"github.com/micro/go-config/source/env"
	"github.com/micro/go-config/source/file"
)

const DefaultApiListenAddress = "0.0.0.0:1337"
const DefaultLogLevel = "ERROR"
const DefaultShutdownTimeoutSeconds = 30

type Configuration struct {
	ApiListenAddress       string `json:"api_listen_addr" yaml:"api_listen_addr"`
	LogLevel               string `json:"log_level" yaml:"log_level"`
	ShutdownTimeoutSeconds int    `json:"shutdown_timeout_seconds" yaml:"shutdown_timeout_seconds"`
	Tls                    struct {
		Cert string `json:"cert" yaml:"key"`
		Key  string `json:"key" yaml:"key"`
	}
	Uaa struct {
		Address  string `json:"address" yaml:"address"`
		Username string `json:"username" yaml:"username"`
		Password string `json:"password" yaml:"password"`
	} `json:"uaa" yaml:"uaa"`
}

func GetConfig(configFilePath *string) Configuration {
	var vcfcsConfig Configuration
	vcfcsConfig.ApiListenAddress = DefaultApiListenAddress
	vcfcsConfig.LogLevel = DefaultLogLevel
	vcfcsConfig.ShutdownTimeoutSeconds = DefaultShutdownTimeoutSeconds

	if configFilePath == nil || *configFilePath == "" {
		return vcfcsConfig
	} else {
		conf := config.NewConfig()
		conf.Load(file.NewSource(
			file.WithPath(*configFilePath),
		),
			env.NewSource(env.WithStrippedPrefix("VAULT_CFCS")))
		conf.Scan(&vcfcsConfig)
		return vcfcsConfig
	}
}
