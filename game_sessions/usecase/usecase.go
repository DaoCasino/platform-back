package usecase

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"platform-backend/blockchain"
	"platform-backend/contracts"
	gamesessions "platform-backend/game_sessions"
	"platform-backend/models"
	"platform-backend/subscription"
	"platform-backend/utils"
	"strconv"
	"time"

	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/eoscanada/eos-go/token"
	"github.com/rs/zerolog/log"
)

type GameSessionsUseCase struct {
	bc               *blockchain.Blockchain
	repo             gamesessions.Repository
	contractsRepo    contracts.Repository
	platformContract string
	subsUseCase      subscription.UseCase
}

func NewGameSessionsUseCase(
	bc *blockchain.Blockchain,
	repo gamesessions.Repository,
	contractsRepo contracts.Repository,
	platformContract string,
	subsUseCase subscription.UseCase,
) *GameSessionsUseCase {
	rand.Seed(time.Now().Unix())
	return &GameSessionsUseCase{
		bc:               bc,
		repo:             repo,
		contractsRepo:    contractsRepo,
		platformContract: platformContract,
		subsUseCase:      subsUseCase,
	}
}

// session in game contract(parse only used fields)
type gameSession struct {
	ReqId      eos.Uint64 `json:"req_id"`
	LastUpdate string     `json:"last_update"`
}

// casino error message
type casinoError struct {
	Error string `json:"error"`
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
					log.Warn().Msgf("EXP_CLEAN: transaction error, "+
						"game contract: %s, sessionID: %d, error: %s", game.Contract, session.ReqId, err.Error())
				}
			}
		}
	}
	return nil
}

func (a *GameSessionsUseCase) NewSession(
	ctx context.Context, casino *models.Casino,
	game *models.Game, user *models.User,
	deposit string, actionType uint16,
	actionParams []uint64,
) (*models.GameSession, error) {
	if casino.Meta == nil {
		return nil, gamesessions.ErrCasinoMetaEmpty
	}

	if casino.Meta.ApiURL == "" {
		return nil, gamesessions.ErrCasinoUrlNotDefined
	}

	playerInfo, err := a.contractsRepo.GetPlayerInfo(ctx, user.AccountName)
	if err != nil {
		return nil, err
	}

	// TODO fix after front lib fix
	sessionId := uint64(rand.Uint32())

	txOpts := a.bc.GetTrxOpts()
	if err := txOpts.FillFromChain(a.bc.Api); err != nil {
		return nil, fmt.Errorf("filling tx opts: %s", err)
	}

	asset, err := utils.ToBetAsset(deposit)
	if err != nil {
		return nil, err
	}

	realAsset, bonusAsset, err := a.getAssets(asset, playerInfo, casino.Id)
	if err != nil {
		return nil, err
	}

	var transferAction *eos.Action

	if realAsset.Amount > 0 {
		// Add transfer deposit action
		transferAction, err = a.getTransferAction(user.AccountName, game.Contract, casino.Contract, sessionId, realAsset)
		if err != nil {
			return nil, err
		}
	}

	newGameAction := a.getNewGameAction(bonusAsset, game.Contract, user, sessionId, casino.Id)

	firstGameAction := &eos.Action{
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
			SessionId:    sessionId,
			ActionType:   actionType,
			ActionParams: actionParams,
		}),
	}

	gameSession := &models.GameSession{
		ID:              sessionId,
		Player:          user.AccountName,
		CasinoID:        casino.Id,
		GameID:          game.Id,
		BlockchainSesID: sessionId,
		State:           models.NewGameTrxSent,
		LastOffset:      0,
		Deposit:         asset,
		LastUpdate:      time.Now().Unix(),
		PlayerWinAmount: nil,
		StateBeforeFail: nil,
	}

	if err := a.repo.AddGameSession(ctx, gameSession); err != nil {
		return nil, err
	}

	err = a.repo.AddGameSessionUpdate(ctx, &models.GameSessionUpdate{
		SessionID:  sessionId,
		UpdateType: models.SessionCreatedUpdate,
		Timestamp:  time.Now(),
		Data:       nil,
		Offset:     nil,
	})
	if err != nil {
		return nil, err
	}

	log.Debug().Msgf("Created new session, sessionID: %d", sessionId)

	// no need to wait response, because that can cause race between action monitor event and casino response
	// sometimes action monitor event about created session handles earlier than casino response
	go func() {
		var trx *eos.Transaction
		if transferAction == nil {
			trx = eos.NewTransaction([]*eos.Action{newGameAction, firstGameAction}, txOpts)
		} else {
			trx = eos.NewTransaction([]*eos.Action{transferAction, newGameAction, firstGameAction}, txOpts)
		}
		trxID, err := a.trxByCasino(casino, trx)
		if err != nil {
			log.Info().Msgf("Error while newgame trx, sessionID: %d, error: %s", sessionId, err.Error())

			e := a.repo.UpdateSessionState(ctx, sessionId, models.GameFailed)
			if e != nil {
				log.Error().Msgf("Error while updating state to failed: %s", e.Error())
				return
			}

			failedUpdateData, e := json.Marshal(struct {
				Details string `json:"details"`
			}{
				Details: err.Error(),
			})
			if e != nil {
				log.Error().Msgf("Error while marshal failed update date: %s", e.Error())
				return
			}

			e = a.repo.AddGameSessionUpdate(ctx, &models.GameSessionUpdate{
				SessionID:  sessionId,
				UpdateType: models.GameFailedUpdate,
				Timestamp:  time.Now(),
				Data:       failedUpdateData,
				Offset:     nil,
			})
			if e != nil {
				log.Error().Msgf("Error pushing session failed update: %s", err.Error())
				return
			}

			//TODO move notify logic to AddGameSessionUpdate
			updates, e := a.repo.GetGameSessionUpdates(ctx, sessionId)
			if e != nil {
				log.Error().Msgf("Error fetching session updates: %s", e.Error())
				return
			}
			updateMsgs := make([]*models.GameSessionUpdateMsg, len(updates))
			for i, update := range updates {
				updateMsgs[i] = models.ToGameSessionUpdateMsg(update)
			}
			go a.subsUseCase.Notify(user.AccountName, "session_update", updateMsgs)
			return
		}
		if err = a.repo.AddGameSessionTransaction(ctx, trxID.String(), sessionId, actionType, actionParams); err != nil {
			log.Warn().Msgf("Failed to add transaction to game_transactions_table, "+
				"sessionID: %d, trxID: %s, reason: %s", sessionId, trxID.String(), err.Error())
		}
		log.Info().Msgf("Successfully sent newgame trx, sessionID: %d, trxID: %s", sessionId, trxID.String())
	}()

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

	trxID, err := a.bc.PushTransaction(
		[]*eos.Action{bcAction},
		[]ecc.PublicKey{a.bc.PubKeys.GameAction},
		false,
	)
	if err != nil {
		return err
	}

	log.Info().Msgf("Successfully sent game action trx, sessionID: %d, trxID: %s", sessionId, trxID.String())

	err = a.repo.UpdateSessionState(ctx, sessionId, models.GameActionTrxSent)
	if err != nil {
		log.Debug().Msgf("Failed to update session state, "+
			"reason: %s", err.Error())
		return err
	}
	if err = a.repo.AddGameSessionTransaction(ctx, trxID.String(), sessionId, actionType, actionParams); err != nil {
		log.Warn().Msgf("Failed to add transaction to game_transactions_table, "+
			"sessionID: %d, trxID: %s, reason: %s", sessionId, trxID.String(), err.Error())
		return err
	}
	return nil
}

func (a *GameSessionsUseCase) GameActionWithDeposit(
	ctx context.Context,
	sessionId uint64,
	actionType uint16,
	actionParams []uint64,
	deposit string,
) error {
	gs, err := a.repo.GetGameSession(ctx, sessionId)
	if err != nil {
		return err
	}

	game, err := a.contractsRepo.GetGame(ctx, gs.GameID)
	if err != nil {
		return err
	}

	casino, err := a.contractsRepo.GetCasino(ctx, gs.CasinoID)
	if err != nil {
		return err
	}

	playerInfo, err := a.contractsRepo.GetPlayerInfo(ctx, gs.Player)
	if err != nil {
		return err
	}

	asset, err := utils.ToBetAsset(deposit)
	if err != nil {
		return err
	}

	realAsset, bonusAsset, err := a.getAssets(asset, playerInfo, casino.Id)
	if err != nil {
		return err
	}

	var transferAction *eos.Action
	var depositBonusAction *eos.Action

	if realAsset.Amount > 0 {
		transferAction, err = a.getTransferAction(gs.Player, game.Contract, casino.Contract, gs.ID, realAsset)
		if err != nil {
			return err
		}
	}

	if bonusAsset.Amount > 0 {
		depositBonusAction = &eos.Action{
			Account: eos.AN(game.Contract),
			Name:    eos.ActN("depositbon"),
			Authorization: []eos.PermissionLevel{{
				Actor:      eos.AN(a.platformContract),
				Permission: eos.PN("gameaction"),
			}},
			ActionData: eos.NewActionData(struct {
				SessionId uint64          `json:"ses_id"`
				From      eos.AccountName `json:"from"`
				Quantity  eos.Asset       `json:"quantity"`
			}{
				SessionId: gs.BlockchainSesID,
				From:      eos.AN(gs.Player),
				Quantity:  *bonusAsset,
			}),
		}
	}

	gameAction := &eos.Action{
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

	txOpts := a.bc.GetTrxOpts()
	if err := txOpts.FillFromChain(a.bc.Api); err != nil {
		return fmt.Errorf("filling tx opts: %s", err)
	}

	trxActions := make([]*eos.Action, 0, 3)
	if transferAction != nil {
		trxActions = append(trxActions, transferAction)
	}
	if depositBonusAction != nil {
		trxActions = append(trxActions, depositBonusAction)
	}
	trxActions = append(trxActions, gameAction)

	trx := eos.NewTransaction(trxActions, txOpts)
	trxID, err := a.trxByCasino(casino, trx)
	if err != nil {
		return err
	}

	log.Info().Msgf("Successfully sent game action with deposit trx, sessionID: %d, trxID: %s", sessionId, trxID.String())

	err = a.repo.UpdateSessionState(ctx, sessionId, models.GameActionTrxSent)
	if err != nil {
		log.Debug().Msgf("%s", err.Error())
		return err
	}
	totalDeposit := gs.Deposit.Add(*asset)
	depositValue, tkn, precision := utils.ExtractAssetValueAndSymbol(&totalDeposit)
	err = a.repo.UpdateSessionDeposit(ctx, sessionId, totalDeposit.String(), depositValue, tkn, precision)
	if err != nil {
		log.Error().Msgf("Failed to update session deposit, "+
			"sesID: %d, trxID: %s, reason: %s", sessionId, trxID.String(), err.Error())
		return err
	}
	if err = a.repo.AddGameSessionTransaction(ctx, trxID.String(), sessionId, actionType, actionParams); err != nil {
		log.Warn().Msgf("Failed to add transaction to game_transactions_table, "+
			"sesID: %d, trxID: %s, reason: %s", sessionId, trxID.String(), err.Error())
		return err
	}
	return nil
}

func (a *GameSessionsUseCase) getTransferAction(
	playerName string,
	gameName string,
	casinoName string,
	sessionID uint64,
	amount *eos.Asset,
) (*eos.Action, error) {
	from := eos.AN(playerName)
	to := eos.AN(gameName)

	memo := strconv.FormatUint(sessionID, 10) // IMPORTANT!

	// Add transfer deposit action
	transferAction := token.NewTransfer(from, to, *amount, memo)
	transferAction.Authorization = []eos.PermissionLevel{
		{Actor: from, Permission: eos.PN(casinoName)},
	}

	return transferAction, nil
}

func (a *GameSessionsUseCase) trxByCasino(casino *models.Casino, trx *eos.Transaction) (eos.Checksum256, error) {
	// Add sponsorship to the transaction
	sponsoredTrx, err := a.bc.GetSponsoredTrx(trx)
	if err != nil {
		return nil, err
	}

	// Sign transaction with GameAction and deposit platform keys
	requiredKeys := []ecc.PublicKey{a.bc.PubKeys.GameAction, a.bc.PubKeys.Deposit}
	signedTrx, err := a.bc.Api.Signer.Sign(sponsoredTrx, a.bc.ChainId, requiredKeys...)
	if err != nil {
		return nil, err
	}

	packedTrx, _, err := signedTrx.PackedTransactionAndCFD()
	if err != nil {
		return nil, err
	}
	h := sha256.New()
	_, _ = h.Write(packedTrx)
	trxID := h.Sum(nil)

	toSend, _ := json.Marshal(signedTrx)
	log.Debug().Msgf("Prepared trx for casino, trx_id: %s, trx: %s", hex.EncodeToString(trxID), string(toSend))

	// Send sponsored and signed transaction to casino Backend
	reader := bytes.NewReader(toSend)
	resp, err := http.Post(casino.Meta.ApiURL+"/sign_transaction", "application/json", reader)
	if err != nil {
		log.Debug().Msgf("Casino request error: %s", err.Error())
		return nil, err
	}
	// don't forget to close response body
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := ioutil.ReadAll(resp.Body)
		log.Debug().Msgf("deposit error from casino back, code %s, body: %s", resp.Status, string(errBody))
		//TODO handle error types
		casError := &casinoError{}
		e := json.Unmarshal(errBody, casError)
		if e != nil {
			return nil, fmt.Errorf("unknown casino error")
		}
		return nil, fmt.Errorf("%s", casError.Error)
	}

	return trxID, nil
}

func (a *GameSessionsUseCase) getAssets(asset *eos.Asset, playerInfo *models.PlayerInfo, casinoId uint64) (*eos.Asset, *eos.Asset, error) {
	bonusBalance := &models.BonusBalance{
		Balance: eos.Asset{
			Amount: 0,
			Symbol: asset.Symbol,
		},
		CasinoId: casinoId,
	}

	for _, bb := range playerInfo.BonusBalances {
		if bb.CasinoId == casinoId {
			bonusBalance = bb
			break
		}
	}

	if playerInfo.Balance.Amount+bonusBalance.Balance.Amount < asset.Amount {
		return nil, nil, fmt.Errorf("not enough tokens")
	}

	realAsset := &eos.Asset{Amount: 0, Symbol: asset.Symbol}
	bonusAsset := &eos.Asset{Amount: 0, Symbol: asset.Symbol}

	if playerInfo.Balance.Amount < asset.Amount {
		bonusAsset.Amount = asset.Amount - playerInfo.Balance.Amount
		realAsset.Amount = playerInfo.Balance.Amount
	} else {
		realAsset = asset
	}

	return realAsset, bonusAsset, nil
}

func (a *GameSessionsUseCase) getNewGameAction(
	bonusAsset *eos.Asset, account string,
	user *models.User, sessionId uint64, casinoId uint64,
) *eos.Action {
	var newGameAction *eos.Action

	if bonusAsset.Amount > 0 {
		//Add newgamebon call to the game to the transaction
		newGameAction = &eos.Action{
			Account: eos.AN(account),
			Name:    eos.ActN("newgamebon"),
			Authorization: []eos.PermissionLevel{{
				Actor:      eos.AN(a.platformContract),
				Permission: eos.PN("gameaction"),
			}},
			ActionData: eos.NewActionData(struct {
				SesId       uint64          `json:"ses_id"`
				CasinoID    uint64          `json:"casino_id"`
				From        eos.AccountName `json:"from"`
				Quantity    eos.Asset       `json:"quantity"`
				AffiliateID string          `json:"affiliate_id"`
			}{
				SesId:       sessionId,
				CasinoID:    casinoId,
				From:        eos.AN(user.AccountName),
				Quantity:    *bonusAsset,
				AffiliateID: user.AffiliateID,
			}),
		}
	} else {
		//Add newgame call to the game to the transaction
		newGameAction = &eos.Action{
			Account: eos.AN(account),
			Name:    eos.ActN("newgame"),
			Authorization: []eos.PermissionLevel{{
				Actor:      eos.AN(a.platformContract),
				Permission: eos.PN("gameaction"),
			}},
			ActionData: eos.NewActionData(struct {
				SesId    uint64 `json:"ses_id"`
				CasinoID uint64 `json:"casino_id"`
			}{
				SesId:    sessionId,
				CasinoID: casinoId,
			}),
		}

		// if user has affiliate call "newgameaffl" action with affiliateID
		if user.AffiliateID != "" {
			newGameAction.Name = eos.ActN("newgameaffl")
			newGameAction.ActionData = eos.NewActionData(struct {
				SesId       uint64 `json:"ses_id"`
				CasinoID    uint64 `json:"casino_id"`
				AffiliateID string `json:"affiliate_id"`
			}{
				SesId:       sessionId,
				CasinoID:    casinoId,
				AffiliateID: user.AffiliateID,
			})
		}
	}

	return newGameAction
}
