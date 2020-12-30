package blockchain

import (
	"context"
	"encoding/json"
	"platform-backend/blockchain"
	"platform-backend/contracts"
	"platform-backend/models"
	"strconv"

	"github.com/eoscanada/eos-go"
	"github.com/rs/zerolog/log"
)

type Casino struct {
	Id        eos.Uint64           `json:"id"`
	Contract  string               `json:"contract"`
	Paused    int                  `json:"paused"`
	RsaPubkey string               `json:"rsa_pubkey"`
	Meta      blockchain.ByteArray `json:"meta"`
}

type GameParam struct {
	Type  uint16     `json:"first"`
	Value eos.Uint64 `json:"second"`
}

type CasinoGame struct {
	Id     eos.Uint64  `json:"game_id"`
	Paused int         `json:"paused"`
	Params []GameParam `json:"params"`
}

type BonusBalance struct {
	Player  string    `json:"player"`
	Balance eos.Asset `json:"balance"`
}

type CasinoBlockchainRepo struct {
	bc               *blockchain.Blockchain
	platformContract string
	bonusActive      bool
}

func NewCasinoBlockchainRepo(
	blockchain *blockchain.Blockchain,
	platformContract string,
	bonusActive bool,
) *CasinoBlockchainRepo {
	return &CasinoBlockchainRepo{
		bc:               blockchain,
		platformContract: platformContract,
		bonusActive:      bonusActive,
	}
}

func (r *CasinoBlockchainRepo) AllCasinos(ctx context.Context) ([]*models.Casino, error) {
	resp, err := r.bc.Api.GetTableRows(eos.GetTableRowsRequest{
		Code:  r.platformContract,
		Scope: r.platformContract,
		Table: "casino",
		Limit: 100,
		JSON:  true,
	})

	if err != nil {
		return nil, err
	}

	casinos := make([]*Casino, 100)
	err = resp.JSONToStructs(&casinos)
	if err != nil {
		return nil, err
	}

	ret := make([]*models.Casino, 0)
	for _, casino := range casinos {
		ret = append(ret, toModelCasino(casino))
	}

	return ret, nil
}

func (r *CasinoBlockchainRepo) GetCasino(ctx context.Context, casinoId uint64) (*models.Casino, error) {
	resp, err := r.bc.Api.GetTableRows(eos.GetTableRowsRequest{
		Code:       r.platformContract,
		Scope:      r.platformContract,
		Table:      "casino",
		Limit:      1,
		LowerBound: strconv.FormatUint(casinoId, 10),
		JSON:       true,
	})

	if err != nil {
		return nil, err
	}

	casinos := make([]*Casino, 1)
	err = resp.JSONToStructs(&casinos)
	if err != nil {
		return nil, err
	}

	if len(casinos) == 0 || uint64(casinos[0].Id) != casinoId {
		return nil, contracts.CasinoNotFound
	}

	return toModelCasino(casinos[0]), nil
}

func (r *CasinoBlockchainRepo) GetCasinoGames(ctx context.Context, casinoName string) ([]*models.CasinoGame, error) {
	resp, err := r.bc.Api.GetTableRows(eos.GetTableRowsRequest{
		Code:  casinoName,
		Scope: casinoName,
		Table: "game",
		Limit: 100,
		JSON:  true,
	})

	if err != nil {
		return nil, err
	}

	casinosGames := make([]*CasinoGame, 1)
	err = resp.JSONToStructs(&casinosGames)
	if err != nil {
		return nil, err
	}

	ret := make([]*models.CasinoGame, 0)
	for _, game := range casinosGames {
		ret = append(ret, toModelCasinoGame(game))
	}

	return ret, nil
}

func (r *CasinoBlockchainRepo) GetBonusBalances(casinos []*models.Casino,
	accountName string) ([]*models.BonusBalance, error) {
	if !r.bonusActive {
		return nil, nil
	}

	bonusBalances := make([]*models.BonusBalance, 0, 1)
	for _, casino := range casinos {
		primaryKey := strconv.FormatUint(eos.MustStringToName(accountName), 10)
		resp, err := r.bc.Api.GetTableRows(eos.GetTableRowsRequest{
			Code:       casino.Contract,
			Scope:      casino.Contract,
			Table:      "bonusbalance",
			LowerBound: primaryKey,
			UpperBound: primaryKey,
			Limit:      1,
			JSON:       true,
		})
		if err != nil {
			return nil, err
		}

		bonusBalance := make([]*BonusBalance, 0, 1)
		err = resp.JSONToStructs(&bonusBalance)
		if err != nil {
			return nil, err
		}

		if len(bonusBalance) == 0 {
			continue
		}
		bonusBalances = append(bonusBalances, toModelBonusBalance(bonusBalance[0], casino.Id))
	}
	return bonusBalances, nil
}

func (r *CasinoBlockchainRepo) getTokenContractBalances(tokenContract eos.AccountName,
	playerAccount eos.AccountName) ([]eos.Asset, error) {
	resp, err := r.bc.Api.GetTableRows(eos.GetTableRowsRequest{
		Code:  string(tokenContract),
		Scope: string(playerAccount),
		Table: "accounts",
		JSON:  true,
	})
	if err != nil {
		return nil, err
	}
	var jsonBalances []struct {
		Balance eos.Asset `json:"balance"`
	}
	if err := resp.JSONToStructs(&jsonBalances); err != nil {
		return nil, err
	}
	balances := make([]eos.Asset, len(jsonBalances))
	for i, b := range jsonBalances {
		balances[i] = b.Balance
	}
	return balances, nil
}

type TokenContract struct {
	TokenName string          `json:"token_name"`
	Contract  eos.AccountName `json:"contract"`
}

func (r *CasinoBlockchainRepo) getTokenToContract() (map[string]eos.AccountName, error) {
	// TODO cache the response
	resp, err := r.bc.Api.GetTableRows(eos.GetTableRowsRequest{
		Code:  r.platformContract,
		Scope: r.platformContract,
		Table: "token",
		JSON:  true,
	})
	if err != nil {
		return nil, err
	}
	var jsonTokenContract []TokenContract
	if err := resp.JSONToStructs(&jsonTokenContract); err != nil {
		return nil, err
	}
	tokenContracts := make(map[string]eos.AccountName)
	for _, tc := range jsonTokenContract {
		tokenContracts[tc.TokenName] = tc.Contract
	}
	return tokenContracts, nil
}

func (r *CasinoBlockchainRepo) GetCustomTokenBalances(casinoName string,
	accountName string) (map[string]eos.Asset, error) {
	resp, err := r.bc.Api.GetTableRows(eos.GetTableRowsRequest{
		Code:  casinoName,
		Scope: casinoName,
		Table: "token",
		Limit: 1000,
		JSON:  true,
	})
	if err != nil {
		return nil, err
	}
	var jsonCasinoSupportedTokens []struct {
		Name   string `json:"token_name"`
		Paused int    `json:"paused"`
	}
	err = resp.JSONToStructs(&jsonCasinoSupportedTokens)
	if err != nil {
		return nil, err
	}
	casinoSupportedTokens := make(map[string]bool)
	for _, token := range jsonCasinoSupportedTokens {
		if token.Name != contracts.CoreSymbol {
			casinoSupportedTokens[token.Name] = true
		}
	}
	playerBalances := make(map[string]eos.Asset)
	seen := make(map[eos.AccountName]bool)
	tokenToContract, err := r.getTokenToContract()
	if err != nil {
		return nil, err
	}
	for platformToken, tokenContract := range tokenToContract {
		if _, found := casinoSupportedTokens[platformToken]; !found {
			continue
		}
		if _, skip := seen[tokenContract]; skip {
			// different tokens can map to the same contract account
			continue
		}
		seen[tokenContract] = true
		balances, err := r.getTokenContractBalances(tokenContract, eos.AN(accountName))
		if err != nil {
			return nil, err
		}
		for _, b := range balances {
			token := b.Symbol.Symbol
			if _, supported := casinoSupportedTokens[token]; supported {
				playerBalances[token] = b
			}
		}
	}
	return playerBalances, nil
}

func toModelCasino(c *Casino) *models.Casino {
	meta := &models.CasinoMeta{}
	err := json.Unmarshal(c.Meta, meta)
	if err != nil {
		log.Warn().Msgf("invalid casino meta, setting null, ID: %d, err: %s", c.Id, err.Error())
		// set null meta if invalid json
		meta = nil
	}

	return &models.Casino{
		Id:       uint64(c.Id),
		Contract: c.Contract,
		Paused:   !(c.Paused == 0),
		Meta:     meta,
	}
}

func toModelCasinoGame(game *CasinoGame) *models.CasinoGame {
	params := make([]models.GameParam, 0)

	for _, param := range game.Params {
		params = append(params, models.GameParam{
			Type:  param.Type,
			Value: uint64(param.Value),
		})
	}

	return &models.CasinoGame{
		Id:     uint64(game.Id),
		Paused: !(game.Paused == 0),
		Params: params,
	}
}

func toModelBonusBalance(bonusBalance *BonusBalance, casinoID uint64) *models.BonusBalance {
	return &models.BonusBalance{
		Balance:  bonusBalance.Balance,
		CasinoId: casinoID,
	}
}
