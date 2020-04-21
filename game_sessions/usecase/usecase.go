package usecase

import (
	"bytes"
	"context"
	"fmt"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/eoscanada/eos-go/token"
	"math/rand"
	"net/http"
	"platform-backend/blockchain"
	"platform-backend/casino"
	"platform-backend/game_sessions"
	"platform-backend/models"
)

type GameSessionsUseCase struct {
	bc               *blockchain.Blockchain
	repo             gamesessions.Repository
	casinoRepo       casino.Repository
	platformContract string
	casinoBackendUrl string
}

func NewGameSessionsUseCase(bc *blockchain.Blockchain, repo gamesessions.Repository, casinoRepo casino.Repository, platformContract string, casinoBackendUrl string) *GameSessionsUseCase {
	return &GameSessionsUseCase{bc: bc, repo: repo, casinoRepo: casinoRepo, platformContract: platformContract, casinoBackendUrl: casinoBackendUrl}
}

func (a *GameSessionsUseCase) NewSession(ctx context.Context, Casino *models.Casino, Game *models.Game, User *models.User, Deposit string) (*models.GameSession, error) {
	api := a.bc.Api

	sessionId := rand.Uint64()

	from := eos.AccountName(User.AccountName)
	to := eos.AccountName(Game.Contract)
	quantity, err := eos.NewEOSAssetFromString(Deposit)
	if err != nil {
		return nil, err
	}
	memo := string(sessionId) // IMPORTANT!

	txOpts := &eos.TxOptions{}
	if err := txOpts.FillFromChain(api); err != nil {
		panic(fmt.Errorf("filling tx opts: %s", err))
	}

	// Add transfer deposit action
	transferAction := token.NewTransfer(from, to, quantity, memo)
	transferAction.Authorization = []eos.PermissionLevel{
		{Actor: from, Permission: eos.PN(Casino.Contract)},
	}

	// Add newgame call to the game to the transaction
	newGameAction := &eos.Action{
		Account: eos.AN(Game.Contract),
		Name:    eos.ActN("newgame"),
		Authorization: []eos.PermissionLevel{
			{Actor: from, Permission: eos.PN("game")},
		},
		ActionData: eos.NewActionData(struct {
			ReqId    uint64 `json:"req_id"`
			CasinoID uint64 `json:"casino_id"`
		}{ReqId: sessionId, CasinoID: Casino.Id}),
	}

	trx := eos.NewTransaction([]*eos.Action{transferAction, newGameAction}, txOpts)

	// Add sponsorship to the transaction
	sponsoredTrx, err := a.bc.GetSponsoredTrx(trx)
	if err != nil {
		return nil, err
	}

	// Sign transaction with GameAction and Deposit platform keys
	requiredKeys := []ecc.PublicKey{a.bc.PubKeys.GameAction, a.bc.PubKeys.Deposit}
	signedTrx, err := api.Signer.Sign(sponsoredTrx, a.bc.ChainId, requiredKeys...)
	if err != nil {
		return nil, err
	}

	// Send sponsored and signed transaction to Casino Backend
	reader := bytes.NewReader([]byte(signedTrx.String()))
	_, err = http.Post(a.casinoBackendUrl+"/sign_transaction", "application/json", reader)
	if err != nil {
		return nil, err
	}

	gameSession := &models.GameSession{
		ID:              sessionId,
		Player:          User.AccountName,
		CasinoID:        Casino.Id,
		GameID:          Game.Id,
		BlockchainSesID: sessionId,
		State:           0,
	}

	if err := a.repo.AddGameSession(ctx, gameSession); err != nil {
		return nil, err
	}

	return gameSession, nil
}
func (a *GameSessionsUseCase) HasGameSession(ctx context.Context, id uint64) (bool, error) {
	return a.repo.HasGameSession(ctx, id)
}
func (a *GameSessionsUseCase) GetGameSession(ctx context.Context, id uint64) (*models.GameSession, error) {
	return a.repo.GetGameSession(ctx, id)
}
