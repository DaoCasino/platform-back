package repository

import (
	"context"
	"platform-backend/contracts"
	"platform-backend/contracts/repository/cached"
	"platform-backend/contracts/repository/mock"
	"platform-backend/models"
	"platform-backend/utils"
	"testing"
	"time"

	"github.com/eoscanada/eos-go"
	"github.com/stretchr/testify/assert"
)

func getInitialData() (eos.AccountResp, models.Game, models.Casino,
	[]*models.CasinoGame, []*models.BonusBalance, map[string]eos.Asset) {
	testRawAccount := eos.AccountResp{
		AccountName: "testuser",
		CoreLiquidBalance: eos.Asset{
			Amount: 100000,
			Symbol: eos.Symbol{
				Precision: 4,
				Symbol:    "BET",
			},
		},
		Permissions: []eos.Permission{{
			PermName: "testcasino",
			Parent:   "active",
		}},
	}
	testGame := models.Game{
		Id:        0,
		Contract:  "testgame",
		ParamsCnt: 2,
		Paused:    0,
		Meta:      nil,
	}
	testCasino := models.Casino{
		Id:       0,
		Contract: "testcasino",
		Paused:   false,
		Meta:     nil,
	}
	testCasinoGames := []*models.CasinoGame{{
		Id:     0,
		Paused: false,
		Params: []models.GameParam{{
			Type:  0,
			Value: 0,
		}, {
			Type:  1,
			Value: 1,
		}},
	}}
	testBonusBalance := []*models.BonusBalance{{
		Balance:  eos.Asset{Amount: 10000, Symbol: eos.Symbol{Precision: 4, Symbol: utils.DAOBetAssetSymbol}},
		CasinoId: 0,
	}}

	eth, _ := eos.NewAssetFromString("4.0000 ETH")
	btc, _ := eos.NewAssetFromString("0.42 BTC")

	testCustomTokenBalances := map[string]eos.Asset{
		"ETH": eth,
		"BTC": btc,
	}
	return testRawAccount, testGame, testCasino, testCasinoGames, testBonusBalance, testCustomTokenBalances
}

func TestCacheInitialization(t *testing.T) {
	cacheTTL := int64(1)
	mockRepo := mock.NewMockedListingRepo()

	testRawAccount, testGame, testCasino, testCasinoGames,
		testBonusBalances, testCustomTokenBalances := getInitialData()

	mockRepo.AddRawAccount(&testRawAccount)
	mockRepo.AddGame(&testGame)
	mockRepo.AddCasino(&testCasino)
	mockRepo.AddCasinoGames(testCasino.Contract, testCasinoGames)

	cachedMockRepo, err := cached.NewCachedListingRepo(mockRepo, cacheTTL)
	assert.NoError(t, err)

	mockRepo.
		On("GetBonusBalances", []models.Casino{testCasino}, string(testRawAccount.AccountName)).
		Return(testBonusBalances, nil)
	mockRepo.
		On("GetCustomTokenBalances", testCasino.Contract, string(testRawAccount.AccountName)).
		Return(testCustomTokenBalances, nil)

	playerInfo, err := cachedMockRepo.GetPlayerInfo(context.Background(), string(testRawAccount.AccountName))
	assert.NoError(t, err)
	assert.Equal(t, testRawAccount.CoreLiquidBalance, playerInfo.Balance)
	assert.Equal(t, []*models.Casino{&testCasino}, playerInfo.LinkedCasinos)
	assert.Equal(t, testBonusBalances, playerInfo.BonusBalances)

	game, err := cachedMockRepo.GetGame(context.Background(), 0)
	assert.NoError(t, err)
	assert.Equal(t, testGame, *game)

	casino, err := cachedMockRepo.GetCasino(context.Background(), 0)
	assert.NoError(t, err)
	assert.Equal(t, testCasino, *casino)

	casinoGames, err := cachedMockRepo.GetCasinoGames(context.Background(), testCasino.Contract)
	assert.NoError(t, err)
	assert.Equal(t, testCasinoGames, casinoGames)
}

func TestCacheAddNewItem(t *testing.T) {
	cacheTTL := int64(1)
	mockRepo := mock.NewMockedListingRepo()

	testRawAccount, testGame, testCasino, testCasinoGames, _, _ := getInitialData()

	mockRepo.AddRawAccount(&testRawAccount)
	mockRepo.AddGame(&testGame)
	mockRepo.AddCasino(&testCasino)
	mockRepo.AddCasinoGames(testCasino.Contract, testCasinoGames)

	cachedMockRepo, err := cached.NewCachedListingRepo(mockRepo, cacheTTL)
	assert.NoError(t, err)

	newTestGame := models.Game{
		Id:        1,
		Contract:  "testgame2",
		ParamsCnt: 2,
		Paused:    0,
		Meta:      nil,
	}
	mockRepo.AddGame(&newTestGame)

	// cache haven't updated yet
	_, err = cachedMockRepo.GetGame(context.Background(), newTestGame.Id)
	assert.EqualError(t, err, contracts.GameNotFound.Error())

	time.Sleep(time.Millisecond * 1010)
	// trigger cache updating, but still return old value
	_, err = cachedMockRepo.GetGame(context.Background(), newTestGame.Id)
	assert.EqualError(t, err, contracts.GameNotFound.Error())

	time.Sleep(time.Millisecond * 100)
	// cache updated, returns new value
	_, err = cachedMockRepo.GetGame(context.Background(), newTestGame.Id)
	assert.NoError(t, err)
}

func TestCacheRemoveItem(t *testing.T) {
	cacheTTL := int64(1)
	mockRepo := mock.NewMockedListingRepo()

	testRawAccount, testGame, testCasino, testCasinoGames, _, _ := getInitialData()

	mockRepo.AddRawAccount(&testRawAccount)
	mockRepo.AddGame(&testGame)
	mockRepo.AddCasino(&testCasino)
	mockRepo.AddCasinoGames(testCasino.Contract, testCasinoGames)

	// add one more game
	newTestGame := models.Game{
		Id:        1,
		Contract:  "testgame2",
		ParamsCnt: 2,
		Paused:    0,
		Meta:      nil,
	}
	mockRepo.AddGame(&newTestGame)

	cachedMockRepo, err := cached.NewCachedListingRepo(mockRepo, cacheTTL)
	assert.NoError(t, err)

	_, err = cachedMockRepo.GetGame(context.Background(), newTestGame.Id)
	assert.NoError(t, err)

	mockRepo.RemoveGame(newTestGame.Id)
	time.Sleep(time.Millisecond * 1010)
	// cache updating triggered, but still returns old value
	_, err = cachedMockRepo.GetGame(context.Background(), newTestGame.Id)
	assert.NoError(t, err)

	time.Sleep(time.Millisecond * 100)
	// cache updated, returns new game should be removed
	_, err = cachedMockRepo.GetGame(context.Background(), newTestGame.Id)
	assert.EqualError(t, err, contracts.GameNotFound.Error())
}

func TestCacheUpdateItem(t *testing.T) {
	cacheTTL := int64(1)
	mockRepo := mock.NewMockedListingRepo()

	testRawAccount, testGame, testCasino, testCasinoGames, _, _ := getInitialData()

	mockRepo.AddRawAccount(&testRawAccount)
	mockRepo.AddGame(&testGame)
	mockRepo.AddCasino(&testCasino)
	mockRepo.AddCasinoGames(testCasino.Contract, testCasinoGames)

	cachedMockRepo, err := cached.NewCachedListingRepo(mockRepo, cacheTTL)
	assert.NoError(t, err)

	// update game
	updatedTestGame := testGame
	updatedTestGame.Paused = 1
	mockRepo.UpdateGame(&updatedTestGame)

	// cache hasn't updated yet
	game, err := cachedMockRepo.GetGame(context.Background(), testGame.Id)
	assert.NoError(t, err)
	assert.Equal(t, testGame, *game)

	time.Sleep(time.Millisecond * 1010)
	// update triggered, but not updated yet
	game, err = cachedMockRepo.GetGame(context.Background(), testGame.Id)
	assert.NoError(t, err)
	assert.Equal(t, testGame, *game)

	time.Sleep(time.Millisecond * 100)
	game, err = cachedMockRepo.GetGame(context.Background(), testGame.Id)
	assert.NoError(t, err)
	assert.Equal(t, updatedTestGame, *game)
}

func TestSortedGames(t *testing.T) {
	cacheTTL := int64(1)
	mockRepo := mock.NewMockedListingRepo()

	testRawAccount, testGame, testCasino, testCasinoGames, _, _ := getInitialData()

	mockRepo.AddRawAccount(&testRawAccount)
	mockRepo.AddGame(&testGame)
	mockRepo.AddCasino(&testCasino)
	mockRepo.AddCasinoGames(testCasino.Contract, testCasinoGames)

	// add one more game
	newTestGame := models.Game{
		Id:        1,
		Contract:  "testgame2",
		ParamsCnt: 2,
		Paused:    0,
		Meta:      nil,
	}
	mockRepo.AddGame(&newTestGame)

	cachedMockRepo, err := cached.NewCachedListingRepo(mockRepo, cacheTTL)
	assert.NoError(t, err)

	games, err := cachedMockRepo.AllGames(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, testGame, *games[0])
	assert.Equal(t, newTestGame, *games[1])
}

func TestSortedCasinos(t *testing.T) {
	cacheTTL := int64(1)
	mockRepo := mock.NewMockedListingRepo()

	testRawAccount, testGame, testCasino, testCasinoGames, _, _ := getInitialData()

	mockRepo.AddRawAccount(&testRawAccount)
	mockRepo.AddGame(&testGame)
	mockRepo.AddCasino(&testCasino)
	mockRepo.AddCasinoGames(testCasino.Contract, testCasinoGames)

	// add one more game
	newTestCasino := models.Casino{
		Id:       1,
		Contract: "testcasino2",
		Paused:   false,
		Meta:     nil,
	}
	mockRepo.AddCasino(&newTestCasino)

	cachedMockRepo, err := cached.NewCachedListingRepo(mockRepo, cacheTTL)
	assert.NoError(t, err)

	casinos, err := cachedMockRepo.AllCasinos(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, testCasino, *casinos[0])
	assert.Equal(t, newTestCasino, *casinos[1])
}
