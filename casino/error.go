package casino

import "errors"

var (
	CasinoNotFound = errors.New("casino not found")
	GameNotFound   = errors.New("game not found")
)
