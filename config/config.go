package config

import (
	"encoding/json"
	"io/ioutil"
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
	KeyPath string `json:"keyPath"`
}

type Config struct {
	DbConfig            DbConfig            `json:"dbConfig"`
	AmcConfig           AmcConfig           `json:"amcConfig"`
	CasinoBackendConfig CasinoBackendConfig `json:"casinoBackendConfig"`
	BlockchainConfig    BlockchainConfig    `json:"blockchainConfig"`
	AuthConfig          AuthConfig          `json:"authConfig"`
	SignidiceConfig     SignidiceConfig     `json:"signidice"`
	LogLevel            string              `json:"loglevel"`
	Port                string              `json:"port"`
}

func FromFile(fileName string) (*Config, error) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	appConfig := &Config{}
	err = json.Unmarshal(data, appConfig)
	if err != nil {
		return nil, err
	}
	return appConfig, nil
}
