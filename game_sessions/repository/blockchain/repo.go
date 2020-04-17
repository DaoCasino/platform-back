package blockchain

import (
	"platform-backend/blockchain"
)

type GameSessionsBCRepo struct {
	bc *blockchain.Blockchain
}

func NewGameSessionsBCRepo(bc *blockchain.Blockchain) *GameSessionsBCRepo {
	return &GameSessionsBCRepo{bc: bc}
}
