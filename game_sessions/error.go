package gamesessions

import "errors"

var (
	ErrGameSessionNotFound = errors.New("game session not found")
	ErrFirstGameActionNotFound = errors.New("first game action not found")
)
