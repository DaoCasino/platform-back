package blockchain

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/eoscanada/eos-go/token"
	"math/rand"
	"net/http"
	"platform-backend/models"
)

func (r *GameSessionsBCRepo) HasGameSession(ctx context.Context, id uint64) (bool, error) {
	return false, errors.New("not implemented")
}

func (r *GameSessionsBCRepo) GetGameSession(ctx context.Context, id uint64) (*models.GameSession, error) {
	return nil, errors.New("not implemented")
}

func (r *GameSessionsBCRepo) AddGameSession(ctx context.Context, casino *models.Casino, game *models.Game, user *models.User, deposit string) (*models.GameSession, error) {
	api := r.bc.Api

	sessionId := rand.Uint64()

	from := eos.AccountName(user.AccountName)
	to := eos.AccountName(game.Contract)
	quantity, err := eos.NewEOSAssetFromString(deposit)
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
		{Actor: from, Permission: eos.PN(casino.Contract)},
	}

	// Add newgame call to the game to the transaction
	newGameAction := &eos.Action{
		Account: eos.AN(game.Contract),
		Name:    eos.ActN("newgame"),
		Authorization: []eos.PermissionLevel{
			{Actor: from, Permission: eos.PN("game")},
		},
		ActionData: eos.NewActionData(struct {
			ReqId    uint64 `json:"req_id"`
			CasinoID uint64 `json:"casino_id"`
		}{ReqId: sessionId, CasinoID: casino.Id}),
	}

	trx := eos.NewTransaction([]*eos.Action{transferAction, newGameAction}, txOpts)

	// Add sponsorship to the transaction
	sponsoredTrx, err := r.bc.GetSponsoredTrx(trx)
	if err != nil {
		return nil, err
	}

	// Sign transaction with GameAction and Deposit platform keys
	requiredKeys := []ecc.PublicKey{r.bc.PubKeys.GameAction, r.bc.PubKeys.Deposit}
	signedTrx, err := api.Signer.Sign(sponsoredTrx, r.bc.ChainId, requiredKeys...)
	if err != nil {
		return nil, err
	}

	packedTrx, err := signedTrx.Pack(eos.CompressionNone)
	if err != nil {
		return nil, err
	}

	// Send sponsored and signed transaction to Casino Backend
	reader := bytes.NewReader(packedTrx.PackedTransaction)
	_, err = http.Post(r.casinoBackendUrl, "application/json", reader)
	if err != nil {
		return nil, err
	}

	return &models.GameSession{
		ID:              sessionId,
		Player:          user.AccountName,
		CasinoID:        casino.Id,
		GameID:          game.Id,
		BlockchainSesID: sessionId,
		State:           0,
	}, nil
}

func (r *GameSessionsBCRepo) DeleteGameSession(ctx context.Context, id uint64) error {
	return errors.New("not implemented")
}
