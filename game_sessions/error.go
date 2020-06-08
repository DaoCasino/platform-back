package gamesessions

import "errors"

var (
	ErrGameSessionNotFound = errors.New("game session not found")
	ErrFirstGameActionNotFound = errors.New("first game action not found")
	ErrCasinoMetaEmpty = errors.New("casino meta is empty")
	ErrCasinoUrlNotDefined = errors.New("casino api url not defined")
)
