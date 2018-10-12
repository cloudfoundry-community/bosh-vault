package config

import (
	"github.com/micro/go-config"
	"github.com/micro/go-config/source/file"
)

const DefaultApiListenAddress = "0.0.0.0:1337"
const DefaultLogLevel = "DEBUG"
const DefaultShutdownTimeoutSeconds = 30

type Configuration struct {
	ApiListenAddress       string `json:"api_listen_addr" yaml:"api_listen_addr"`
	LogLevel               string `json:"log_level" yaml:"log_level"`
	ShutdownTimeoutSeconds int    `json:"shutdown_timeout_seconds" yaml:"shutdown_timeout_seconds"`
	TlsCertPath            string `json:"tls_cert_path" yaml:"tls_cert_path"`
	TlsKeyPath             string `json:"tls_key_path" yaml:"tls_key_path"`
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
		))
		conf.Scan(&vcfcsConfig)
		return vcfcsConfig
	}
}
