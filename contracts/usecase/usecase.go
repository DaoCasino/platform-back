package usecase

import (
	"context"
	"errors"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/rs/zerolog/log"
	"platform-backend/blockchain"
)

type ContractsUseCase struct {
	bc *blockchain.Blockchain
}

func NewContractsUseCase(bc *blockchain.Blockchain) *ContractsUseCase {
	return &ContractsUseCase{bc: bc}
}

func (c *ContractsUseCase) SendBonusToNewPlayer(ctx context.Context, accountName string, casinoName string) error {
	if casinoName == "" {
		return errors.New("casino name is not defined")
	}

	action := &eos.Action{
		Account: eos.AN(casinoName),
		Name:    eos.ActN("newplayer"),
		Authorization: []eos.PermissionLevel{{
			Actor:      eos.AN(c.bc.PlatformAccountName),
			Permission: eos.PN("gameaction"),
		}},
		ActionData: eos.NewActionData(struct {
			PlayerAccount eos.AccountName `json:"player_account"`
		}{
			PlayerAccount: eos.AN(accountName),
		}),
	}

	trxID, err := c.bc.PushTransaction([]*eos.Action{action}, []ecc.PublicKey{c.bc.PubKeys.GameAction}, false)
	if err != nil {
		return err
	}

	log.Info().Msgf("Successfully sent newplayer trx to player %s, trxID: %s", accountName, trxID.String())

	return nil
}
