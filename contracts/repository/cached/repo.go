package cached

import (
	"context"
	"github.com/eoscanada/eos-go"
	"github.com/rs/zerolog/log"
	"go.uber.org/atomic"
	"platform-backend/contracts"
	"platform-backend/models"
	"sync"
	"time"
)

type CachedListingRepo struct {
	origRepo contracts.Repository
	cacheTTL int64

	cacheMutex      sync.RWMutex
	lastCacheUpdate time.Time
	cacheUpdating   atomic.Bool

	games       map[uint64]*models.Game
	casinos     map[uint64]*models.Casino
	casinoGames map[string][]*models.CasinoGame
}

func NewCachedListingRepo(origRepo contracts.Repository, cacheTTL int64) (*CachedListingRepo, error) {
	repo := CachedListingRepo{
		origRepo: origRepo,
		cacheTTL: cacheTTL,

		games:       make(map[uint64]*models.Game),
		casinos:     make(map[uint64]*models.Casino),
		casinoGames: make(map[string][]*models.CasinoGame),
	}

	err := repo.initCache()
	if err != nil {
		return nil, err
	}

	return &repo, nil
}

func (r *CachedListingRepo) _refreshCache(needLock bool) error {
	// fetch all data to local variables
	games, err := r.origRepo.AllGames(context.Background())
	if err != nil {
		return err
	}
	casinos, err := r.origRepo.AllCasinos(context.Background())
	if err != nil {
		return err
	}
	casinoGames := map[string][]*models.CasinoGame{}
	for _, cas := range casinos {
		casGames, err := r.origRepo.GetCasinoGames(context.Background(), cas.Contract)
		if err != nil {
			return err
		}
		casinoGames[cas.Contract] = casGames
	}

	// if need lock get mutex only after data fetched
	if needLock {
		defer r.cacheMutex.Unlock()
		r.cacheMutex.Lock()
	}

	// clean old data
	r.games = make(map[uint64]*models.Game)
	r.casinos = make(map[uint64]*models.Casino)
	r.casinoGames = make(map[string][]*models.CasinoGame)

	// update cached data
	for _, game := range games {
		r.games[game.Id] = game
	}
	for _, casino := range casinos {
		r.casinos[casino.Id] = casino
	}
	for casinoName, casGames := range casinoGames {
		r.casinoGames[casinoName] = casGames
	}

	r.lastCacheUpdate = time.Now()

	return nil
}

// synchronously obtain initial cache data
func (r *CachedListingRepo) initCache() error {
	// get full lock to initialize cache
	defer r.cacheMutex.Unlock()
	r.cacheMutex.Lock()

	// no need to lock because we already have full write lock
	return r._refreshCache(false)
}

// NOTE: function creates go routine for cache updating
func (r *CachedListingRepo) tryUpdateCache() {
	// if cache already updating just skip
	if r.cacheUpdating.Load() {
		return
	}

	// check cache update time
	if r.lastCacheUpdate.Unix()+r.cacheTTL > time.Now().Unix() {
		return
	}

	// refresh cache in new go routine
	go func() {
		r.cacheUpdating.Store(true)
		defer r.cacheUpdating.Store(false)

		// update data with soft write lock
		err := r._refreshCache(true)
		if err != nil {
			log.Warn().Msgf("Listing cached updating fail: %s", err.Error())
			return
		}
		log.Debug().Msgf("Listing cache successfully updated")
	}()
	log.Debug().Msgf("Started listing cache updating")
}

func (r *CachedListingRepo) AllCasinos(ctx context.Context) ([]*models.Casino, error) {
	defer func() {
		// try to update cache on demand
		r.tryUpdateCache()
		r.cacheMutex.RUnlock()
	}()
	r.cacheMutex.RLock()

	// preallocate array with known capacity
	ret := make([]*models.Casino, 0, len(r.casinos))
	for _, casino := range r.casinos {
		casCopy := *casino
		ret = append(ret, &casCopy)
	}

	return ret, nil
}

func (r *CachedListingRepo) GetCasino(ctx context.Context, casinoId uint64) (*models.Casino, error) {
	defer func() {
		// try to update cache on demand
		r.tryUpdateCache()
		r.cacheMutex.RUnlock()
	}()
	r.cacheMutex.RLock()

	if casino, ok := r.casinos[casinoId]; ok {
		casCopy := *casino
		return &casCopy, nil
	}

	return nil, contracts.CasinoNotFound
}

func (r *CachedListingRepo) GetCasinoGames(ctx context.Context, casinoName string) ([]*models.CasinoGame, error) {
	defer func() {
		// try to update cache on demand
		r.tryUpdateCache()
		r.cacheMutex.RUnlock()
	}()
	r.cacheMutex.RLock()

	if casGames, ok := r.casinoGames[casinoName]; ok {
		return casGames, nil
	}

	return nil, contracts.CasinoNotFound
}

func (r *CachedListingRepo) AllGames(ctx context.Context) ([]*models.Game, error) {
	defer func() {
		// try to update cache on demand
		r.tryUpdateCache()
		r.cacheMutex.RUnlock()
	}()
	r.cacheMutex.RLock()

	// preallocate array with known capacity
	ret := make([]*models.Game, 0, len(r.games))
	for _, game := range r.games {
		gameCopy := *game
		ret = append(ret, &gameCopy)
	}

	return ret, nil
}

func (r *CachedListingRepo) GetGame(ctx context.Context, gameId uint64) (*models.Game, error) {
	defer func() {
		// try to update cache on demand
		r.tryUpdateCache()
		r.cacheMutex.RUnlock()
	}()
	r.cacheMutex.RLock()

	if game, ok := r.games[gameId]; ok {
		gameCopy := *game
		return &gameCopy, nil
	}

	return nil, contracts.GameNotFound
}

// cached only linked casino
func (r *CachedListingRepo) GetPlayerInfo(ctx context.Context, accountName string) (*models.PlayerInfo, error) {
	rawAccount, err := r.GetRawAccount(accountName)
	if err != nil {
		return nil, err
	}

	casinos, err := r.AllCasinos(ctx)
	if err != nil {
		return nil, err
	}

	info := &models.PlayerInfo{}
	contracts.FillPlayerInfoFromRaw(info, rawAccount, casinos)

	return info, nil
}

// without cache, just fwd to original repo
func (r *CachedListingRepo) GetRawAccount(accountName string) (*eos.AccountResp, error) {
	return r.origRepo.GetRawAccount(accountName)
}
