package localstorage

import "context"

func (r *GameSessionsLocalRepo) AddGameSessionTransaction(
	_ context.Context,
	_ string, _ uint64,
	_ uint16, _ []uint64) error {
	return nil
}
