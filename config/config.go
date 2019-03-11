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
const DefaultUaaAudienceClaim = "config_server"
const DefaultUaaKeyRefreshIntervalSeconds = 86400
const DefaultVaultConnectionTimeoutSeconds = 30
const DefaultVaultMount = "secret"

type Configuration struct {
	Api struct {
		Address      string `json:"address" yaml:"address"`
		DrainTimeout int    `json:"draintimeout" yaml:"draintimeout"`
	} `json:"api" yaml:"api"`
	Log struct {
		Level string `json:"level" yaml:"level"`
	} `json:"log" yaml:"log"`
	Tls struct {
		Cert string `json:"cert" yaml:"key"`
		Key  string `json:"key" yaml:"key"`
	} `json:"tls" yaml:"tls"`
	Redirects []RedirectBlock    `json:"redirects" yaml:"redirects"`
	Uaa       UaaConfiguration   `json:"uaa" yaml:"uaa"`
	Vault     VaultConfiguration `json:"vault" yaml:"vault"`
	Debug     DebugConfiguration `json:"debug" yaml:"debug"`
}

type DebugConfiguration struct {
	DisableAuth bool `json:"disable_auth" yaml:"disable_auth"`
}

type UaaConfiguration struct {
	Address               string `json:"address" yaml:"address"`
	Timeout               int    `json:"timeout" yaml:"timeout"`
	Ca                    string `json:"ca" yaml:"ca"`
	SkipVerify            bool   `json:"skipverify" yaml:"skipverify"`
	ExpectedAudienceClaim string `json:"audienceclaim"`
	KeyRefreshInterval    int    `json:"keyrefreshinterval" yaml:"keyrefreshinterval"`
}

type VaultConfiguration struct {
	Address         string `json:"address" yaml:"address"`
	Token           string `json:"token" yaml:"token"`
	Timeout         int    `json:"timeout" yaml:"timeout"`
	Mount           string `json:"mount" yaml:"mount"`
	Ca              string `json:"ca" yaml:"ca"`
	SkipVerify      bool   `json:"skipverify" yaml:"skipverify"`
	RenewalInterval int    `json:"renewalinterval" yaml:"renewalinterval"`
}

type RedirectRule struct {
	Ref      string `json:"ref" yaml:"ref"`
	Redirect string `json:"redirect" yaml:"redirect"`
}

type RedirectBlock struct {
	Type  string             `json:"type" yaml:"type"`
	Vault VaultConfiguration `json:"vault" yaml:"vault"`
	Rules []RedirectRule     `json:"rules" yaml:"rules"`
}

func ParseConfig(configFilePath *string) Configuration {
	var bvConfig Configuration
	bvConfig.Api.Address = DefaultApiListenAddress
	bvConfig.Log.Level = DefaultLogLevel
	bvConfig.Api.DrainTimeout = DefaultShutdownTimeoutSeconds
	bvConfig.Uaa.Timeout = DefaultUaaConnectionTimeoutSeconds
	bvConfig.Uaa.ExpectedAudienceClaim = DefaultUaaAudienceClaim
	bvConfig.Uaa.KeyRefreshInterval = DefaultUaaKeyRefreshIntervalSeconds
	bvConfig.Vault.Timeout = DefaultVaultConnectionTimeoutSeconds
	bvConfig.Vault.Mount = DefaultVaultMount

	if configFilePath == nil || *configFilePath == "" {
		return bvConfig
	} else {
		conf := config.NewConfig()
		// ok to ignore errors on loading because of defaulting behavior
		_ = conf.Load(file.NewSource(
			file.WithPath(*configFilePath)),
			env.NewSource(env.WithStrippedPrefix("BV")),
		)
		_ = conf.Scan(&bvConfig)
		return bvConfig
	}
}
