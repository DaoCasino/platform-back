package mock

import (
	"context"
	"errors"
	"github.com/eoscanada/eos-go"
	"github.com/stretchr/testify/mock"
	"platform-backend/contracts"
	"platform-backend/models"
)

type MockedListingRepo struct {
	rawAccounts map[string]*eos.AccountResp
	games       map[uint64]*models.Game
	casinos     map[uint64]*models.Casino
	casinoGames map[string][]*models.CasinoGame
	mock.Mock
}

func NewMockedListingRepo() *MockedListingRepo {
	return &MockedListingRepo{
		rawAccounts: make(map[string]*eos.AccountResp),
		games:       make(map[uint64]*models.Game),
		casinos:     make(map[uint64]*models.Casino),
		casinoGames: make(map[string][]*models.CasinoGame),
	}
}

func (r *MockedListingRepo) AllCasinos(ctx context.Context) ([]*models.Casino, error) {
	// preallocate array with known capacity
	ret := make([]*models.Casino, 0, len(r.casinos))
	for _, casino := range r.casinos {
		casCopy := *casino
		ret = append(ret, &casCopy)
	}

	return ret, nil
}

func (r *MockedListingRepo) GetCasino(ctx context.Context, casinoId uint64) (*models.Casino, error) {
	if casino, ok := r.casinos[casinoId]; ok {
		casCopy := *casino
		return &casCopy, nil
	}

	return nil, contracts.CasinoNotFound
}

func (r *MockedListingRepo) GetCasinoGames(ctx context.Context, casinoName string) ([]*models.CasinoGame, error) {
	if casGames, ok := r.casinoGames[casinoName]; ok {
		return casGames, nil
	}

	return nil, contracts.CasinoNotFound
}

func (r *MockedListingRepo) AllGames(ctx context.Context) ([]*models.Game, error) {
	// preallocate array with known capacity
	ret := make([]*models.Game, 0, len(r.games))
	for _, game := range r.games {
		gameCopy := *game
		ret = append(ret, &gameCopy)
	}

	return ret, nil
}

func (r *MockedListingRepo) GetGame(ctx context.Context, gameId uint64) (*models.Game, error) {
	if game, ok := r.games[gameId]; ok {
		gameCopy := *game
		return &gameCopy, nil
	}

	return nil, contracts.GameNotFound
}

func (r *MockedListingRepo) GetPlayerInfo(ctx context.Context, accountName string) (*models.PlayerInfo, error) {
	rawAccount, err := r.GetRawAccount(accountName)
	if err != nil {
		return nil, err
	}

	casinos, err := r.AllCasinos(ctx)
	if err != nil {
		return nil, err
	}

	info := &models.PlayerInfo{}
	contracts.FillPlayerInfoFromRaw(info, rawAccount, casinos, nil)

	return info, nil
}

func (r *MockedListingRepo) GetRawAccount(accountName string) (*eos.AccountResp, error) {
	if account, ok := r.rawAccounts[accountName]; ok {
		return account, nil
	}
	return nil, errors.New("account not found")
}

func (r *MockedListingRepo) GetBonusBalances(casinos []*models.Casino, accountName string) ([]*models.BonusBalance, error) {
	valueCasinos := make([]models.Casino, 0)
	for _, casino := range casinos {
		valueCasinos = append(valueCasinos, *casino)
	}
	args := r.Called(valueCasinos, accountName)
	return args.Get(0).([]*models.BonusBalance), args.Error(1)
}

func (r *MockedListingRepo) AddRawAccount(account *eos.AccountResp) {
	r.rawAccounts[string(account.AccountName)] = account
}

func (r *MockedListingRepo) AddGame(game *models.Game) {
	r.games[game.Id] = game
}

func (r *MockedListingRepo) UpdateGame(game *models.Game) {
	r.games[game.Id] = game
}

func (r *MockedListingRepo) RemoveGame(gameId uint64) {
	delete(r.games, gameId)
}

func (r *MockedListingRepo) AddCasino(casino *models.Casino) {
	r.casinos[casino.Id] = casino
	r.casinoGames[casino.Contract] = make([]*models.CasinoGame, 0)
}

func (r *MockedListingRepo) AddCasinoGames(casinoName string, casinoGames []*models.CasinoGame) {
	r.casinoGames[casinoName] = casinoGames
}
