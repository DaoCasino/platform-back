package models

type ContractSession struct {
	ReqId      uint64           `json:"req_id"`
	CasinoId   uint64           `json:"casino_id"`
	SesSeq     uint64           `json:"ses_seq"`
	Player     string           `json:"player"`
	State      GameSessionState `json:"state"`
	Deposit    string           `json:"deposit"`
	Digest     string           `json:"digest"`
	LastUpdate string           `json:"last_update"`
	LasMaxWin  string           `json:"last_max_win"`
}
