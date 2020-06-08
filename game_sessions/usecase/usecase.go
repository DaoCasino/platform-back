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
	"io/ioutil"
	"math/rand"
	"net/http"
	"platform-backend/blockchain"
	"platform-backend/contracts"
	"platform-backend/game_sessions"
	"platform-backend/models"
	"strconv"
	"time"
)

type GameSessionsUseCase struct {
	bc               *blockchain.Blockchain
	repo             gamesessions.Repository
	contractsRepo    contracts.Repository
	platformContract string
}

func NewGameSessionsUseCase(
	bc *blockchain.Blockchain,
	repo gamesessions.Repository,
	contractsRepo contracts.Repository,
	platformContract string,
) *GameSessionsUseCase {
	rand.Seed(time.Now().Unix())
	return &GameSessionsUseCase{
		bc:               bc,
		repo:             repo,
		contractsRepo:    contractsRepo,
		platformContract: platformContract,
	}
}

// session in game contract(parse only used fields)
type gameSession struct {
	ReqId      eos.Uint64 `json:"req_id"`
	LastUpdate string     `json:"last_update"`
}

func (a *GameSessionsUseCase) CleanExpiredSessions(
	ctx context.Context,
	maxLastUpdate time.Duration,
) error {
	games, err := a.contractsRepo.AllGames(ctx)
	if err != nil {
		return err
	}
	for _, game := range games {
		resp, err := a.bc.Api.GetTableRows(eos.GetTableRowsRequest{
			Code:  game.Contract,
			Scope: game.Contract,
			Table: "session",
			Limit: 1000,
			JSON:  true,
		})
		if err != nil {
			return err
		}

		sessions := &[]gameSession{}
		err = resp.JSONToStructs(sessions)
		if err != nil {
			return err
		}

		for _, session := range *sessions {
			lastUpdate, err := time.Parse("2006-01-02T15:04:05.000", session.LastUpdate)
			if err != nil {
				return err
			}
			if time.Now().After(lastUpdate.Add(maxLastUpdate)) {
				// Session is expired!

				txOpts := a.bc.GetTrxOpts()
				if err := txOpts.FillFromChain(a.bc.Api); err != nil {
					return err
				}
				closeAction := &eos.Action{
					Account: eos.AN(game.Contract),
					Name:    eos.ActN("close"),
					Authorization: []eos.PermissionLevel{
						{Actor: eos.AN(a.platformContract), Permission: eos.PN("gameaction")},
					},
					ActionData: eos.NewActionData(struct {
						ReqId uint64 `json:"req_id"`
					}{ReqId: uint64(session.ReqId)}),
				}
				tx := eos.NewTransaction([]*eos.Action{closeAction}, txOpts)
				notSigned := eos.NewSignedTransaction(tx)

				requiredKeys := []ecc.PublicKey{a.bc.PubKeys.GameAction}
				signedTrx, err := a.bc.Api.Signer.Sign(notSigned, a.bc.ChainId, requiredKeys...)
				if err != nil {
					return err
				}
				packedTrx, err := signedTrx.Pack(eos.CompressionNone)
				if err != nil {
					return err
				}
				if _, err := a.bc.Api.PushTransaction(packedTrx); err != nil {
					// Do not return error, because it can be caused by bug in contract
					// So just log it and ignore
					log.Warn().Msgf("EXP_CLEAN: transaction error: %s", err)
				}
			}
		}
	}
	return nil
}

func (a *GameSessionsUseCase) NewSession(
	ctx context.Context, casino *models.Casino,
	game *models.Game, user *models.User,
	deposit string,
) (*models.GameSession, error) {
	api := a.bc.Api

	if casino.Meta == nil {
		return nil, gamesessions.ErrCasinoMetaEmpty
	}

	if casino.Meta.ApiURL == "" {
		return nil, gamesessions.ErrCasinoUrlNotDefined
	}

	// TODO fix after front lib fix
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
			{Actor: eos.AN(a.platformContract), Permission: eos.PN("gameaction")},
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
	resp, err := http.Post(casino.Meta.ApiURL+"/sign_transaction", "application/json", reader)
	if err != nil {
		log.Debug().Msgf("%s", err.Error())
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		errBody, _ := ioutil.ReadAll(resp.Body)
		log.Error().Msgf("deposit error from casino back, code %s, body: %s", resp.Status, string(errBody))
		return nil, errors.New("casino error")
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

	game, err := a.contractsRepo.GetGame(ctx, gs.GameID)
	if err != nil {
		return err
	}

	bcAction := &eos.Action{
		Account: eos.AN(game.Contract),
		Name:    eos.ActN("gameaction"),
		Authorization: []eos.PermissionLevel{{
			Actor:      eos.AN(a.platformContract),
			Permission: eos.PN("gameaction"),
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

	_, err = a.bc.PushTransaction(
		[]*eos.Action{bcAction},
		[]ecc.PublicKey{a.bc.PubKeys.GameAction},
		false,
	)

	if err != nil {
		return err
	}

	err = a.repo.UpdateSessionState(ctx, sessionId, models.GameActionTrxSent)
	if err != nil {
		log.Debug().Msgf("%s", err.Error())
		return err
	}

	return nil
}
