package usecase

import (
	"bytes"
	"context"
	"encoding/json"
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

func NewGameSessionsUseCase(bc *blockchain.Blockchain, repo gamesessions.Repository, casinoRepo casino.Repository, platformContract string, casinoBackendUrl string) *GameSessionsUseCase {
	rand.Seed(time.Now().Unix())
	return &GameSessionsUseCase{bc: bc, repo: repo, casinoRepo: casinoRepo, platformContract: platformContract, casinoBackendUrl: casinoBackendUrl}
}

func (a *GameSessionsUseCase) NewSession(ctx context.Context, Casino *models.Casino, Game *models.Game, User *models.User, Deposit string) (*models.GameSession, error) {
	api := a.bc.Api

	sessionId := uint64(rand.Uint32())

	from := eos.AccountName(User.AccountName)
	to := eos.AccountName(Game.Contract)
	quantity, err := eos.NewFixedSymbolAssetFromString(eos.Symbol{Precision: 4, Symbol: "BET"}, Deposit)
	if err != nil {
		return nil, err
	}

	memo := strconv.FormatUint(sessionId, 10) // IMPORTANT!

	txOpts := &eos.TxOptions{}
	if err := txOpts.FillFromChain(api); err != nil {
		panic(fmt.Errorf("filling tx opts: %s", err))
	}

	// Add transfer deposit action
	transferAction := token.NewTransfer(from, to, quantity, memo)
	transferAction.Authorization = []eos.PermissionLevel{
		{Actor: from, Permission: eos.PN(Casino.Contract)},
	}

	//Add newgame call to the game to the transaction
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

	toSend, _ := json.Marshal(signedTrx)
	log.Debug().Msgf("Trx to send: %s", string(toSend))

	// Send sponsored and signed transaction to Casino Backend
	reader := bytes.NewReader(toSend)
	_, err = http.Post(a.casinoBackendUrl+"/sign_transaction", "application/json", reader)
	if err != nil {
		log.Debug().Msgf("%s", err.Error())
		return nil, err
	}

	gameSession := &models.GameSession{
		ID:              sessionId,
		Player:          User.AccountName,
		CasinoID:        Casino.Id,
		GameID:          Game.Id,
		BlockchainSesID: sessionId,
		State:           models.NewGameTrxSent,
		LastOffset:      0,
	}

	if err := a.repo.AddGameSession(ctx, gameSession); err != nil {
		return nil, err
	}

	return gameSession, nil
}

func (a *GameSessionsUseCase) GameAction(ctx context.Context, sessionId uint64, actionType uint16, actionParams []uint32) error {
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
			ActionParams []uint32 `json:"params"`
		}{
			SessionId:    gs.BlockchainSesID,
			ActionType:   actionType,
			ActionParams: actionParams,
		}),
	}

	trxOpts := &eos.TxOptions{}
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

	resp, err := a.bc.Api.PushTransaction(packedTrx)
	if err != nil {
		log.Debug().Msgf("%s", err.Error())
		return err
	}

	err = a.repo.UpdateSessionState(ctx, sessionId, models.GameActionTrxSent)
	if err != nil {
		log.Debug().Msgf("%s", err.Error())
		return err
	}

	log.Debug().Msgf("Game action trx, resp code: %d, blk num: %d", resp.StatusCode, resp.BlockNum)
	return nil
}

func (a *GameSessionsUseCase) HasGameSession(ctx context.Context, id uint64) (bool, error) {
	return a.repo.HasGameSession(ctx, id)
}
func (a *GameSessionsUseCase) GetGameSession(ctx context.Context, id uint64) (*models.GameSession, error) {
	return a.repo.GetGameSession(ctx, id)
}
