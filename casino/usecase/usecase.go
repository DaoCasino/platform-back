package usecase

import (
	"context"
	"github.com/eoscanada/eos-go"
	"github.com/rs/zerolog/log"
	"platform-backend/blockchain"
	"platform-backend/casino"
	"platform-backend/game_sessions"
)

type CasinoUseCase struct {
	casinoRepo      casino.Repository
	gameSessionRepo game_sessions.Repository
	bc              *blockchain.Blockchain
	gameKeyBag      *eos.KeyBag
}

func NewCasinoUseCase(
	casinoRepo casino.Repository,
	gsRepo game_sessions.Repository,
	bc *blockchain.Blockchain,
	gameKey string,
) *CasinoUseCase {
	gameKeyBag := eos.NewKeyBag()
	_ = gameKeyBag.Add(gameKey)

	return &CasinoUseCase{
		casinoRepo:      casinoRepo,
		gameSessionRepo: gsRepo,
		bc:              bc,
		gameKeyBag:      gameKeyBag,
	}
}

type GameActionData struct {
	SessionId    uint64   `json:"ses_id"`
	ActionType   uint16   `json:"type"`
	ActionParams []uint32 `json:"params"`
}

func (c *CasinoUseCase) GameAction(ctx context.Context, sessionId uint64, actionType uint16, actionParams []uint32) error {
	gs, err := c.gameSessionRepo.GetGameSession(ctx, sessionId)
	if err != nil {
		return err
	}

	game, err := c.casinoRepo.GetGame(ctx, gs.GameID)
	if err != nil {
		return err
	}

	bcAction := eos.Action{
		Account: eos.AN(game.Contract),
		Name:    eos.ActN("gameaction"),
		Authorization: []eos.PermissionLevel{{
			Actor:      eos.AN(gs.Player),
			Permission: eos.PN("game"),
		}},
		ActionData: eos.NewActionData(GameActionData{
			SessionId:    gs.BlockchainSesID,
			ActionType:   actionType,
			ActionParams: actionParams,
		}),
	}

	trxOpts := eos.TxOptions{}
	err = trxOpts.FillFromChain(c.bc.Api)
	if err != nil {
		log.Err(err)
		return err
	}

	sponsoredTrx, err := c.bc.GetSponsoredTrx(eos.NewTransaction([]*eos.Action{&bcAction}, &trxOpts))
	if err != nil {
		log.Err(err)
		return err
	}

	keys, _ := c.gameKeyBag.AvailableKeys()
	signedTrx, err := c.gameKeyBag.Sign(sponsoredTrx, c.bc.ChainID, keys[0])
	if err != nil {
		log.Err(err)
		return err
	}

	packedTrx, err := signedTrx.Pack(eos.CompressionNone)
	if err != nil {
		log.Err(err)
		return err
	}

	resp, err := c.bc.Api.PushTransaction(packedTrx)
	if err != nil {
		log.Err(err)
		return err
	}

	log.Debug().Msgf("Game action trx, resp code: %s, blk num: %d", resp.StatusCode, resp.BlockNum)
	return nil
}
