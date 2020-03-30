package config

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	DbConfig struct {
		Host string `json:"host"`
		Port string `json:"port"`
	} `json:"dbConfig"`
	LogLevel string `json:"loglevel"`
	Port     string `json:"port"`
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
