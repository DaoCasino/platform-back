package models

import (
	"github.com/eoscanada/eos-go"
	"time"
)

type GameState struct {
	GameId               uint64    `json:"game_id"`
	Balance              eos.Asset `json:"balance"`
	LastClaimTime        time.Time `json:"last_claim_type"`
	ActiveSessionsAmount uint64    `json:"active_sessions_amount"`
	ActiveSessionsSum    eos.Asset `json:"active_sessions_sum"`
}
