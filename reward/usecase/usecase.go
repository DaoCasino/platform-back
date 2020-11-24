package usecase

import (
	"context"
	"errors"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/rs/zerolog/log"
	"platform-backend/blockchain"
	"platform-backend/contracts"
	"platform-backend/models"
	"time"
)

type RewardUseCase struct {
	contractRepo         contracts.Repository
	bc                   *blockchain.Blockchain
	signidiceAccountName string
	rewardInterval       time.Duration
}

func NewRewardUseCase(
	contractRepo contracts.Repository,
	bc *blockchain.Blockchain,
	signidiceAccountName string,
	rewardInterval int,
) *RewardUseCase {
	return &RewardUseCase{
		contractRepo:         contractRepo,
		bc:                   bc,
		signidiceAccountName: signidiceAccountName,
		rewardInterval:       time.Duration(rewardInterval) * time.Second,
	}
}

func (r *RewardUseCase) RewardGameDevs(ctx context.Context) error {
	casinos, err := r.contractRepo.AllCasinos(ctx)
	if err != nil {
		log.Debug().Msgf("get all casinos error: %s", err.Error())
		return err
	}

	games, err := r.contractRepo.AllGames(ctx)
	if err != nil {
		log.Debug().Msgf("get all games error: %s", err.Error())
		return err
	}

	for _, c := range casinos {
		gameStates, err := r.contractRepo.GetCasinoGamesState(ctx, c.Contract)
		if err != nil {
			log.Debug().Msgf("get casino games state error: %s", err.Error())
			return err
		}

		for _, gs := range gameStates {
			if !time.Now().After(gs.LastClaimTime.Add(r.rewardInterval)) {
				continue
			}
			gameName, err := getGameName(gs.GameId, games)
			if err != nil {
				log.Debug().Msgf("get game name error: %s", err.Error())
			}

			if err := r.sendClaimProfit(c.Contract, gameName); err != nil {
				log.Debug().Msgf("send claim profit error: %s", err.Error())
				return err
			}
		}
	}

	return nil
}

func (r *RewardUseCase) sendClaimProfit(casinoName string, gameName string) error {
	action := &eos.Action{
		Account: eos.AN(casinoName),
		Name:    eos.ActN("claimprofit"),
		Authorization: []eos.PermissionLevel{
			{Actor: eos.AN(r.signidiceAccountName), Permission: eos.PN("active")},
		},
		ActionData: eos.NewActionData(struct {
			GameAccount eos.AccountName `json:"game_account"`
		}{
			GameAccount: eos.AN(gameName),
		}),
	}

	trxID, err := r.bc.PushTransaction(
		[]*eos.Action{action},
		[]ecc.PublicKey{r.bc.PubKeys.SigniDice},
		false,
	)
	if err != nil {
		return err
	}

	log.Info().Msgf("Successfully sended reward to gameName: %s, casinoName: %s, trxID: %s",
		gameName, casinoName, trxID.String())

	return nil
}

func getGameName(gameId uint64, games []*models.Game) (string, error) {
	for _, g := range games {
		if g.Id == gameId {
			return g.Contract, nil
		}
	}

	return "", errors.New("game is not found")
}
