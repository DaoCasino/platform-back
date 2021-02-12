package server

import (
	"context"
	"github.com/DaoCasino/platform-action-monitor-client"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"os/signal"
	affiliateStatsRepo "platform-backend/affiliatestats/repository/http"
	"platform-backend/auth"
	authPgRepo "platform-backend/auth/repository/postgres"
	authUC "platform-backend/auth/usecase"
	"platform-backend/blockchain"
	cashbackRepo "platform-backend/cashback/repository/postgres"
	cashbackUC "platform-backend/cashback/usecase"
	"platform-backend/config"
	"platform-backend/contracts"
	contractsBcRepo "platform-backend/contracts/repository/blockchain"
	contractsCachedRepo "platform-backend/contracts/repository/cached"
	contractsUC "platform-backend/contracts/usecase"
	"platform-backend/db"
	"platform-backend/eventprocessor"
	gameSessionPgRepo "platform-backend/game_sessions/repository/postgres"
	gameSessionUC "platform-backend/game_sessions/usecase"
	"platform-backend/logger"
	referralsRepo "platform-backend/referrals/repository/postgres"
	referralsUC "platform-backend/referrals/usecase"
	"platform-backend/repositories"
	"platform-backend/server/api"
	"platform-backend/server/session_manager"
	smLocalRepo "platform-backend/server/session_manager/repository/localstorage"
	signidiceUC "platform-backend/signidice/usecase"
	subscriptionUc "platform-backend/subscription/usecase"
	"platform-backend/usecases"
	"reflect"
	"strconv"
	"time"
)

type JsonResponse = map[string]interface{}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type LogoutRequest struct {
	AccessToken string `json:"accessToken"`
}

type AuthRequest struct {
	TmpToken    string `json:"tmpToken"`
	CasinoName  string `json:"casinoName"`
	AffiliateID string `json:"affiliateID"`
}

type OptOutRequest struct {
	AccessToken string `json:"accessToken"`
}

type SetEthAddrRequest struct {
	AccessToken string `json:"accessToken"`
	EthAddress  string `json:"ethAddress"`
}

type App struct {
	httpHandler http.Handler
	config      *config.Config
	wsUpgrader  websocket.Upgrader
	wsApi       *api.WsApi

	smRepo         session_manager.Repository
	uRepo          auth.UserRepository
	eventProcessor *eventprocessor.EventProcessor
	useCases       *usecases.UseCases
	events         chan *eventlistener.EventMessage

	developmentMode bool
}

const (
	PrometheusPrefix = "platformback_"
	ServiceName      = "platform"
)

func NewApp(config *config.Config) (*App, error) {
	logger.InitLogger(config.LogLevel)

	// Create prometheus things
	commonBuckets := prometheus.LinearBuckets(0, 5, 200)
	registry := prometheus.NewRegistry()
	registerer := prometheus.WrapRegistererWithPrefix(PrometheusPrefix, registry)

	registerer.MustRegister(prometheus.NewGoCollector())

	err := db.InitDB(context.Background(), &config.Db, registerer)
	if err != nil {
		log.Fatal().Msgf("Database init error, %s", err.Error())
		return nil, err
	}

	bc, err := blockchain.Init(&config.Blockchain)
	if err != nil {
		log.Fatal().Msgf("Blockchain init error, %s", err.Error())
		return nil, err
	}

	var contractRepo contracts.Repository
	contractRepo = contractsBcRepo.NewCasinoBlockchainRepo(bc, config.Blockchain.Contracts.Platform, config.ActiveFeatures.Bonus)

	// use cached contract repo if cache enabled
	if config.Blockchain.ListingCacheTTL > 0 {
		contractRepo, err = contractsCachedRepo.NewCachedListingRepo(contractRepo, config.Blockchain.ListingCacheTTL)
		if err != nil {
			log.Fatal().Msgf("Contracts cached repo creation error, %s", err.Error())
			return nil, err
		}
	}

	gsRepo := gameSessionPgRepo.NewGameSessionsPostgresRepo(db.DbPool)
	smRepo := smLocalRepo.NewLocalRepository(registerer)
	uRepo := authPgRepo.NewUserPostgresRepo(db.DbPool, config.Auth.MaxUserSessions, config.Auth.RefreshTokenTTL)
	refsRepo := referralsRepo.NewReferralPostgresRepo(db.DbPool)
	affStatsRepo := affiliateStatsRepo.NewAffiliateStatsRepo(config.AffiliateStats.Url, config.ActiveFeatures.Referrals)
	cbRepo := cashbackRepo.NewCashbackPostgresRepo(db.DbPool)

	repos := repositories.NewRepositories(
		contractRepo,
		gsRepo,
		affStatsRepo,
	)

	subsUC := subscriptionUc.NewSubscriptionUseCase()
	contractUC := contractsUC.NewContractsUseCase(bc, config.ActiveFeatures.Bonus)
	refsUC := referralsUC.NewReferralsUseCase(refsRepo, config.ActiveFeatures.Referrals)
	cbUC := cashbackUC.NewCashbackUseCase(
		cbRepo,
		affStatsRepo,
		config.Cashback.Ratio,
		config.Cashback.EthToBetRate,
		config.ActiveFeatures.Cashback,
	)

	useCases := usecases.NewUseCases(
		authUC.NewAuthUseCase(
			uRepo,
			smRepo,
			cbRepo,
			contractUC,
			[]byte(config.Auth.JwtSecret),
			config.Auth.AccessTokenTTL,
			config.Auth.RefreshTokenTTL,
			config.Auth.WalletURL,
			config.Auth.WalletClientID,
			config.Auth.WalletClientSecret,
		),
		gameSessionUC.NewGameSessionsUseCase(
			bc,
			repos.GameSession,
			repos.Contracts,
			config.Blockchain.Contracts.Platform,
			subsUC,
		),
		signidiceUC.NewSignidiceUseCase(
			bc,
			config.Blockchain.Contracts.Platform,
			config.Signidice.AccountName,
			config.Signidice.Key,
		),
		subsUC,
		refsUC,
		cbUC,
	)

	events := make(chan *eventlistener.EventMessage)

	// Hack for development mode, just set DEV_MODE env to enable
	_, devMode := os.LookupEnv("DEV_MODE")

	app := &App{
		config: config,
		wsUpgrader: websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
			return true
		}},
		smRepo:          smRepo,
		uRepo:           uRepo,
		eventProcessor:  eventprocessor.New(repos, bc, useCases, registerer),
		useCases:        useCases,
		wsApi:           api.NewWsApi(useCases, repos, registerer),
		events:          events,
		developmentMode: devMode,
	}

	wsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wsClientHandler(app, w, r)
	})

	authHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHandler(app, w, r)
	})

	refreshTokensHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		refreshTokensHandler(app, w, r)
	})

	logoutHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logoutHandler(app, w, r)
	})

	optOutHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		optOutHandler(app, w, r)
	})

	setEthAddrHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setEthAddrHandler(app, w, r)
	})

	requestDurationHistograms := make(map[string]*prometheus.HistogramVec)

	requestDurationsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			t := time.Now()
			elapsed := t.Sub(start)
			code := reflect.Indirect(reflect.ValueOf(w)).FieldByName("status").Int()
			requestDurationHistograms[r.RequestURI].WithLabelValues(strconv.FormatInt(code, 10)).Observe(float64(elapsed.Milliseconds()))
		})
	}

	// Create router and add handlers
	r := mux.NewRouter()
	r.Use(requestDurationsMiddleware)

	addHistogramVec := func(path string) string {
		fullPath := "/" + path
		requestDurationHistograms[fullPath] = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_" + path + "_ms",
			Buckets: commonBuckets,
		}, []string{"response_code"})
		return fullPath
	}

	handle := func(path string, handler http.Handler) *mux.Route {
		return r.Handle(addHistogramVec(path), handler)
	}

	handleFunc := func(path string, f func(http.ResponseWriter, *http.Request)) *mux.Route {
		return r.HandleFunc(addHistogramVec(path), f)
	}

	handle("connect", wsHandler)
	handleFunc("auth", authHandler)
	handleFunc("logout", logoutHandler)
	handleFunc("refresh_token", refreshTokensHandler)
	handleFunc("optout", optOutHandler)
	handleFunc("set_eth_addr", setEthAddrHandler)
	handleFunc("ping", pingHandler)
	handleFunc("who", whoHandler)
	handle("metrics", promhttp.InstrumentMetricHandler(
		registerer, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}),
	))

	for _, hist := range requestDurationHistograms {
		registerer.MustRegister(hist)
	}

	app.httpHandler = cors.Default().Handler(r)

	log.Info().Msg("App created")

	return app, nil
}

func startSessionsCleaner(a *App, ctx context.Context) error {
	interval := a.config.SessionsCleaner.Interval
	if interval <= 0 {
		log.Info().Msg("Sessions cleaner is disabled")
		<-ctx.Done()
		return nil
	}

	maxLastUpdate := time.Duration(a.config.SessionsCleaner.MaxLastUpdate) * time.Second

	log.Info().Msg("Sessions cleaner is started")
	clean := func() error {
		log.Info().Msg("Sessions cleaner is cleaning sessions...")
		if err := a.useCases.GameSession.CleanExpiredSessions(ctx, maxLastUpdate); err != nil {
			return err
		}
		log.Info().Msgf("Old sessions were cleaned!")
		return nil
	}
	if err := clean(); err != nil {
		log.Error().Msgf("Sessions clean error: %s", err.Error())
	}

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	for {
		select {
		case <-ticker.C:
			if err := clean(); err != nil {
				log.Error().Msgf("Sessions clean error: %s", err.Error())
			}
		case <-ctx.Done():
			ticker.Stop()
			log.Info().Msg("Sessions cleaner is stopped")
			return nil
		}
	}
}

func startAuthSessionsCleaner(a *App, ctx context.Context) error {
	interval := a.config.Auth.CleanerInterval
	if interval <= 0 {
		log.Info().Msg("Auth sessions cleaner is disabled")
		<-ctx.Done()
		return nil
	}

	log.Info().Msg("Auth sessions cleaner is started")
	clean := func() error {
		log.Info().Msg("Auth sessions cleaner is cleaning sessions...")
		if err := a.uRepo.InvalidateOldSessions(ctx); err != nil {
			return err
		}
		log.Info().Msgf("Old auth sessions were cleaned!")
		return nil
	}
	if err := clean(); err != nil {
		return err
	}

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	for {
		select {
		case <-ticker.C:
			if err := clean(); err != nil {
				return err
			}
		case <-ctx.Done():
			ticker.Stop()
			log.Info().Msg("Auth Sessions cleaner is stopped")
			return nil
		}
	}
}

func startHttpServer(a *App, ctx context.Context) error {
	srv := &http.Server{Addr: ":" + a.config.Port, Handler: a.httpHandler}
	log.Info().Msgf("Server is starting on %s port", a.config.Port)

	if !a.config.ActiveFeatures.Bonus {
		log.Info().Msg("Bonus feature is disabled")
	}
	if !a.config.ActiveFeatures.Referrals {
		log.Info().Msg("Referrals feature is disabled")
	}

	go func() {
		<-ctx.Done()
		timeoutCtx, shutdown := context.WithTimeout(context.Background(), 5*time.Second)
		_ = srv.Shutdown(timeoutCtx)
		shutdown()
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Error().Msgf("ListenAndServe(): %v", err)
		return err
	}

	log.Info().Msgf("Server is stopped")

	return nil
}

func startAmc(a *App, ctx context.Context) error {
	listener := eventlistener.NewEventListener(a.config.Amc.Url, a.events)
	// setup reconnection options
	listener.ReconnectionAttempts = a.config.Amc.ReconnectionAttempts
	listener.ReconnectionDelay = time.Duration(a.config.Amc.ReconnectionDelay)

	log.Info().Msgf("Connecting to the action monitor on %s", a.config.Amc.Url)

	if a.config.LogLevel == "debug" {
		eventlistener.EnableDebugLogging()
	}

	// set auth token
	listener.SetToken(a.config.Amc.Token)

	go listener.Run(ctx)

	if ok, err := listener.BatchSubscribe(eventprocessor.GetEventsToSubscribe(), 0); err != nil || !ok {
		log.Fatal().Msgf("Action monitor subscribe to events, error: %v", err)
	}

	log.Info().Msgf("Subscribed to all events!")

	for {
		select {
		case <-ctx.Done():
			log.Info().Msgf("Action monitor client is stopped")
			return nil
		case eventMessage, ok := <-a.events:
			if !ok {
				return nil
			}
			for _, event := range eventMessage.Events {
				a.eventProcessor.Process(ctx, event)
			}
		}
	}
}

// Should log errors by itself
func (a *App) Run() error {
	runCtx, cancelRun := context.WithCancel(context.Background())
	errGroup, runCtx := errgroup.WithContext(runCtx)

	errGroup.Go(func() error {
		defer cancelRun()
		return startHttpServer(a, runCtx)
	})
	errGroup.Go(func() error {
		defer cancelRun()
		return startAmc(a, runCtx)
	})
	errGroup.Go(func() error {
		defer cancelRun()
		return startSessionsCleaner(a, runCtx)
	})
	errGroup.Go(func() error {
		defer cancelRun()
		return startAuthSessionsCleaner(a, runCtx)
	})

	errGroup.Go(func() error {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt)
		select {
		case <-runCtx.Done():
			return nil
		case <-quit:
			cancelRun()
		}
		return nil
	})

	return errGroup.Wait()
}
