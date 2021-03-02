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

type SessionsCleaner struct {
	Interval      int `json:"interval"`
	MaxLastUpdate int `json:"maxLastUpdate"`
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
	DisableSponsor  bool  `json:"disableSponsor"`
	TrxPushAttempts int   `default:"5" json:"trxPushAttempts"`
	ListingCacheTTL int64 `json:"listingCacheTTL"`
}

type AuthConfig struct {
	JwtSecret          string   `json:"jwtSecret"`
	AccessTokenTTL     int64    `json:"accessTokenTTL"`
	RefreshTokenTTL    int64    `json:"refreshTokenTTL"`
	MaxUserSessions    int64    `default:"20" json:"maxUserSessions"`
	CleanerInterval    int64    `default:"600" json:"cleanerInterval"`
	WalletURL          string   `json:"walletUrl"`
	WalletClientID     int64    `json:"walletClientID"`
	WalletClientSecret string   `json:"walletClientSecret"`
	TestAccounts       []string `json:"testAccounts"`
}

// Action monitor config
type AmcConfig struct {
	Url                  string `json:"url"`
	ReconnectionAttempts int    `default:"5" json:"reconnectionAttempts"`
	ReconnectionDelay    int    `default:"5" json:"reconnectionDelay"`
	Token                string `json:"token"`
}

type SignidiceConfig struct {
	AccountName string `json:"accountName"`
	Key         string `json:"key"`
}

type AffiliateStatsConfig struct {
	Url string `json:"url"`
}

type ActiveFeaturesConfig struct {
	Bonus     bool `default:"true" json:"bonus"`
	Referrals bool `default:"true" json:"referrals"`
	Cashback  bool `default:"true" json:"cashback"`
}

type CashbackConfig struct {
	Ratio        float64 `json:"ratio"`
	EthToBetRate float64 `json:"eth_to_bet_rate"`
}

type Config struct {
	Db              DbConfig             `json:"db"`
	Amc             AmcConfig            `json:"amc"`
	SessionsCleaner SessionsCleaner      `json:"sessionsCleaner"`
	Blockchain      BlockchainConfig     `json:"blockchain"`
	Auth            AuthConfig           `json:"auth"`
	Signidice       SignidiceConfig      `json:"signidice"`
	AffiliateStats  AffiliateStatsConfig `json:"affiliateStats"`
	ActiveFeatures  ActiveFeaturesConfig `json:"activeFeatures"`
	Cashback        CashbackConfig       `json:"cashback"`
	LogLevel        string               `json:"loglevel"`
	Port            string               `json:"port"`
}

func Read(fileName string) (*Config, error) {
	appConfig := &Config{}
	data, err := ioutil.ReadFile(fileName)
	if err == nil {
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
