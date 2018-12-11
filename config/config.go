package config

import (
	"github.com/micro/go-config"
	"github.com/micro/go-config/source/env"
	"github.com/micro/go-config/source/file"
)

const DefaultApiListenAddress = "0.0.0.0:1337"
const DefaultLogLevel = "ERROR"
const DefaultShutdownTimeoutSeconds = 30
const DefaultUaaConnectionTimeoutSeconds = 10
const DefaultVaultConnectionTimeoutSeconds = 30
const DefaultVaultPrefix = "secret"

type Configuration struct {
	ApiListenAddress       string `json:"api_listen_addr" yaml:"api_listen_addr"`
	LogLevel               string `json:"log_level" yaml:"log_level"`
	ShutdownTimeoutSeconds int    `json:"shutdown_timeout_seconds" yaml:"shutdown_timeout_seconds"`
	Vault                  struct {
		Address string `json:"address" yaml:"address"`
		Token   string `json:"token" yaml:"token"`
		Timeout int    `json:"timeout" yaml:"timeout"`
		Prefix  string `json:"prefix" yaml:"prefix"`
	} `json:"vault" yaml:"vault"`
	Tls struct {
		Cert string `json:"cert" yaml:"key"`
		Key  string `json:"key" yaml:"key"`
	} `json:"tls" yaml:"tls"`
	Uaa struct {
		Enabled               bool   `json:"enabled" yaml:"enabled"`
		Address               string `json:"address" yaml:"address"`
		Username              string `json:"username" yaml:"username"`
		Password              string `json:"password" yaml:"password"`
		Timeout               int    `json:"timeout" yaml:"timeout"`
		Ca                    string `json:"ca" yaml:"ca"`
		SkipVerify            bool   `json:"skipverify" yaml:"skipverify"`
		ExpectedAudienceClaim string `json:"audienceclaim"`
	} `json:"uaa" yaml:"uaa"`
}

func GetConfig(configFilePath *string) Configuration {
	var bvConfig Configuration
	bvConfig.ApiListenAddress = DefaultApiListenAddress
	bvConfig.LogLevel = DefaultLogLevel
	bvConfig.ShutdownTimeoutSeconds = DefaultShutdownTimeoutSeconds
	bvConfig.Uaa.Enabled = true
	bvConfig.Uaa.Timeout = DefaultUaaConnectionTimeoutSeconds
	bvConfig.Vault.Timeout = DefaultVaultConnectionTimeoutSeconds
	bvConfig.Vault.Prefix = DefaultVaultPrefix

	if configFilePath == nil || *configFilePath == "" {
		return bvConfig
	} else {
		conf := config.NewConfig()
		conf.Load(file.NewSource(
			file.WithPath(*configFilePath),
		),
			env.NewSource(env.WithStrippedPrefix("BV")))
		conf.Scan(&bvConfig)
		return bvConfig
	}
}
