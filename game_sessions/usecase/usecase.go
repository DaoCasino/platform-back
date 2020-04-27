package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/eoscanada/eos-go/token"
	"github.com/rs/zerolog/log"
	"math/rand"
	"net/http"
	"platform-backend/blockchain"
	"platform-backend/casino"
	"platform-backend/game_sessions"
	"platform-backend/models"
	"strconv"
	"time"
)

type GameSessionsUseCase struct {
	bc               *blockchain.Blockchain
	repo             gamesessions.Repository
	casinoRepo       casino.Repository
	platformContract string
	casinoBackendUrl string
}

func NewGameSessionsUseCase(
	bc *blockchain.Blockchain,
	repo gamesessions.Repository,
	casinoRepo casino.Repository,
	platformContract string,
	casinoBackendUrl string,
) *GameSessionsUseCase {
	rand.Seed(time.Now().Unix())
	return &GameSessionsUseCase{
		bc:               bc,
		repo:             repo,
		casinoRepo:       casinoRepo,
		platformContract: platformContract,
		casinoBackendUrl: casinoBackendUrl,
	}
}

func (a *GameSessionsUseCase) NewSession(
	ctx context.Context, casino *models.Casino,
	game *models.Game, user *models.User,
	deposit string, actionType uint16, actionParams []uint64,
) (*models.GameSession, error) {
	api := a.bc.Api

	sessionId := uint64(rand.Uint32())

	from := eos.AccountName(user.AccountName)
	to := eos.AccountName(game.Contract)
	quantity, err := eos.NewFixedSymbolAssetFromString(eos.Symbol{Precision: 4, Symbol: "BET"}, deposit)
	if err != nil {
		return nil, err
	}

	memo := strconv.FormatUint(sessionId, 10) // IMPORTANT!

	txOpts := a.bc.GetTrxOpts()
	if err := txOpts.FillFromChain(api); err != nil {
		panic(fmt.Errorf("filling tx opts: %s", err))
	}

	// Add transfer deposit action
	transferAction := token.NewTransfer(from, to, quantity, memo)
	transferAction.Authorization = []eos.PermissionLevel{
		{Actor: from, Permission: eos.PN(casino.Contract)},
	}

	//Add newgame call to the game to the transaction
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
	sponsoredTrx, err := a.bc.GetSponsoredTrx(trx)
	if err != nil {
		return nil, err
	}

	// Sign transaction with GameAction and deposit platform keys
	requiredKeys := []ecc.PublicKey{a.bc.PubKeys.GameAction, a.bc.PubKeys.Deposit}
	signedTrx, err := api.Signer.Sign(sponsoredTrx, a.bc.ChainId, requiredKeys...)
	if err != nil {
		return nil, err
	}

	toSend, _ := json.Marshal(signedTrx)
	log.Debug().Msgf("Trx to send: %s", string(toSend))

	// Send sponsored and signed transaction to casino Backend
	reader := bytes.NewReader(toSend)
	resp, err := http.Post(a.casinoBackendUrl+"/sign_transaction", "application/json", reader)
	if err != nil {
		log.Debug().Msgf("%s", err.Error())
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Error().Msgf("deposit error from casino: %s", resp.Status)
		return nil, errors.New("casino error: " + resp.Status)
	}

	gameSession := &models.GameSession{
		ID:              sessionId,
		Player:          user.AccountName,
		CasinoID:        casino.Id,
		GameID:          game.Id,
		BlockchainSesID: sessionId,
		State:           models.NewGameTrxSent,
		LastOffset:      0,
	}

	if err := a.repo.AddGameSession(ctx, gameSession); err != nil {
		return nil, err
	}

	// make game action
	err = a.GameAction(ctx, sessionId, actionType, actionParams)
	if err != nil {
		return nil, err
	}

	return gameSession, nil
}

func (a *GameSessionsUseCase) GameAction(
	ctx context.Context,
	sessionId uint64,
	actionType uint16,
	actionParams []uint64,
) error {
	gs, err := a.repo.GetGameSession(ctx, sessionId)
	if err != nil {
		return err
	}

	game, err := a.casinoRepo.GetGame(ctx, gs.GameID)
	if err != nil {
		return err
	}

	bcAction := &eos.Action{
		Account: eos.AN(game.Contract),
		Name:    eos.ActN("gameaction"),
		Authorization: []eos.PermissionLevel{{
			Actor:      eos.AN(gs.Player),
			Permission: eos.PN("game"),
		}},
		ActionData: eos.NewActionData(struct {
			SessionId    uint64   `json:"ses_id"`
			ActionType   uint16   `json:"type"`
			ActionParams []uint64 `json:"params"`
		}{
			SessionId:    gs.BlockchainSesID,
			ActionType:   actionType,
			ActionParams: actionParams,
		}),
	}

	trxOpts := a.bc.GetTrxOpts()
	err = trxOpts.FillFromChain(a.bc.Api)
	if err != nil {
		log.Debug().Msgf("%s", err.Error())
		return err
	}

	sponsoredTrx, err := a.bc.GetSponsoredTrx(eos.NewTransaction([]*eos.Action{bcAction}, trxOpts))
	if err != nil {
		log.Debug().Msgf("%s", err.Error())
		return err
	}

	signedTrx, err := a.bc.Api.Signer.Sign(sponsoredTrx, a.bc.ChainId, a.bc.PubKeys.GameAction)
	if err != nil {
		log.Debug().Msgf("%s", err.Error())
		return err
	}

	packedTrx, err := signedTrx.Pack(eos.CompressionNone)
	if err != nil {
		log.Debug().Msgf("%s", err.Error())
		return err
	}

	_, err = a.bc.Api.PushTransaction(packedTrx)
	if err != nil {
		log.Debug().Msgf("%s", err.Error())
		return err
	}

	err = a.repo.UpdateSessionState(ctx, sessionId, models.GameActionTrxSent)
	if err != nil {
		log.Debug().Msgf("%s", err.Error())
		return err
	}

	return nil
}
