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

type AuthConfig struct {
	JwtSecret       string `json:"jwtSecret"`
	AccessTokenTTL  int64  `json:"accessTokenTTL"`
	RefreshTokenTTL int64  `json:"refreshTokenTTL"`
}

// Action monitor config
type AmcConfig struct {
	Url string `json:"url"`
}

type Config struct {
	DbConfig   DbConfig   `json:"dbConfig"`
	AmcConfig  AmcConfig  `json:"amcConfig"`
	AuthConfig AuthConfig `json:"authConfig"`
	LogLevel   string     `json:"loglevel"`
	Port       string     `json:"port"`
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
