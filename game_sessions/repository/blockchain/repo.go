package blockchain

import (
	"platform-backend/blockchain"
)

type GameSessionsBCRepo struct {
	bc               *blockchain.Blockchain
	platformContract string
	casinoBackendUrl string
}

func NewGameSessionsBCRepo(bc *blockchain.Blockchain, platformContract string, casinoBackendUrl string) *GameSessionsBCRepo {
	return &GameSessionsBCRepo{bc: bc, platformContract: platformContract, casinoBackendUrl: casinoBackendUrl}
}
