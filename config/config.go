package config

import (
	"encoding/json"
	"github.com/kelseyhightower/envconfig"
	"io/ioutil"
	"os"
)

type DbConfig struct {
	Url          string `json:"url"`
	MaxPoolConns int32  `json:"maxPoolConns"`
	MinPoolConns int32  `json:"minPoolConns"`
}

type BlockchainConfig struct {
	NodeUrl    string `json:"nodeUrl"`
	SponsorUrl string `json:"sponsorUrl"`
	Contracts  struct {
		Platform string `json:"platform"`
	} `json:"contracts"`
	Permissions struct {
		Deposit    string `json:"deposit"`
		GameAction string `json:"gameaction"`
		SigniDice  string `json:"signidice"`
	} `json:"permissions"`
	DisableSponsor bool `json:"disableSponsor"`
}

type AuthConfig struct {
	JwtSecret       string `json:"jwtSecret"`
	AccessTokenTTL  int64  `json:"accessTokenTTL"`
	RefreshTokenTTL int64  `json:"refreshTokenTTL"`
}

// Action monitor config
type AmcConfig struct {
	Url string `json:"url"`
}

type CasinoBackendConfig struct {
	Url string `json:"url"`
}

type SignidiceConfig struct {
	Key string `json:"key"`
}

type Config struct {
	DbConfig   DbConfig            `json:"db"`
	Amc        AmcConfig           `json:"amc"`
	Casino     CasinoBackendConfig `json:"casino"`
	Blockchain BlockchainConfig    `json:"blockchain"`
	Auth       AuthConfig          `json:"auth"`
	Signidice  SignidiceConfig     `json:"signidice"`
	LogLevel   string              `json:"loglevel"`
	Port       string              `json:"port"`
}

func Read(fileName string) (*Config, error) {
	appConfig := &Config{}
	data, err := ioutil.ReadFile(fileName)
	if err == nil  {
		err = json.Unmarshal(data, appConfig)
		if err != nil {
			return nil, err
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	err = envconfig.Process("", appConfig)
	if err != nil {
		return nil, err
	}
	return appConfig, nil
}
