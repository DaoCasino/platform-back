package config

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestConfigRead(t *testing.T) {
	os.Setenv("CASHBACK_RATIO", "0.02")
	os.Setenv("CASHBACK_ETHTOBETRATE", "0.000006400")
	config, err := Read("../config.json")
	assert.NoError(t, err)
	assert.Equal(t, 0.02, config.Cashback.Ratio)
	assert.Equal(t, 0.000006400, config.Cashback.EthToBetRate)
}
