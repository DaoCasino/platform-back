package server

import (
	"context"
	"encoding/json"
	eventlistener "github.com/DaoCasino/platform-action-monitor-client"
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
	"platform-backend/auth"
	authPgRepo "platform-backend/auth/repository/postgres"
	authUC "platform-backend/auth/usecase"
	"platform-backend/blockchain"
	"platform-backend/config"
	casinoBcRepo "platform-backend/contracts/repository/blockchain"
	"platform-backend/db"
	"platform-backend/eventprocessor"
	gameSessionPgRepo "platform-backend/game_sessions/repository/postgres"
	gameSessionUC "platform-backend/game_sessions/usecase"
	"platform-backend/logger"
	"platform-backend/models"
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
	TmpToken string `json:"tmpToken"`
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

const PrometheusPrefix = "platformback_"

func wsClientHandler(app *App, w http.ResponseWriter, r *http.Request) {
	log.Debug().Msgf("New connect request")

	c, err := app.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Err(err)
		return
	}

	log.Info().Msgf("Client with ip %q connected", c.RemoteAddr())

	app.smRepo.AddSession(context.Background(), c, app.wsApi)
}

func authHandler(app *App, w http.ResponseWriter, r *http.Request) {
	var user *models.User
	if app.developmentMode {
		user = &models.User{
			AccountName: "testuserever",
			Email:       "test@user.ever",
		}
	} else {
		var req AuthRequest
		err := json.NewDecoder(r.Body).Decode(&req);
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			log.Debug().Msgf("Http body parse error, %s", err.Error())
			return
		}

		log.Debug().Msgf("New auth request with token %s", req.TmpToken)

		user, err = app.useCases.Auth.ResolveUser(context.Background(), req.TmpToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			log.Debug().Msgf("Token validate error: %s", err.Error())
			return
		}
	}

	refreshToken, accessToken, err := app.useCases.Auth.SignUp(context.Background(), user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Debug().Msgf("SignUp error: %s", err.Error())
		return
	}

	response, err := json.Marshal(JsonResponse{
		"refreshToken": refreshToken,
		"accessToken":  accessToken,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Debug().Msgf("Response marshal error: %s", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(response)
}

func logoutHandler(app *App, w http.ResponseWriter, r *http.Request) {
	log.Debug().Msgf("New logout request")

	var req LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Debug().Msgf("Http body parse error, %s", err.Error())
		return
	}

	err := app.useCases.Auth.Logout(context.Background(), req.AccessToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Debug().Msgf("RefreshToken error: %s", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func refreshTokensHandler(app *App, w http.ResponseWriter, r *http.Request) {
	log.Debug().Msgf("New refresh_token request")

	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Debug().Msgf("Http body parse error, %s", err.Error())
		return
	}

	refreshToken, accessToken, err := app.useCases.Auth.RefreshToken(context.Background(), req.RefreshToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Debug().Msgf("RefreshToken error: %s", err.Error())
		return
	}

	response, err := json.Marshal(JsonResponse{
		"refreshToken": refreshToken,
		"accessToken":  accessToken,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Debug().Msgf("Response marshal error: %s", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(response)
}

func pingHandler(w http.ResponseWriter, _ *http.Request) {
	log.Debug().Msgf("New ping request")
	w.WriteHeader(http.StatusOK)
}

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

	smRepo := smLocalRepo.NewLocalRepository(registerer)

	repos := repositories.NewRepositories(
		casinoBcRepo.NewCasinoBlockchainRepo(bc, config.Blockchain.Contracts.Platform),
		gameSessionPgRepo.NewGameSessionsPostgresRepo(db.DbPool),
	)

	uRepo := authPgRepo.NewUserPostgresRepo(db.DbPool, config.Auth.MaxUserSessions, config.Auth.RefreshTokenTTL)

	subsUC := subscriptionUc.NewSubscriptionUseCase();

	useCases := usecases.NewUseCases(
		authUC.NewAuthUseCase(
			uRepo,
			smRepo,
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
			config.Signidice.Key,
		),
		subsUC,
	)

	events := make(chan *eventlistener.EventMessage)

	// Hack for development mode, just set DEV_MODE env to enable
	_, devMode := os.LookupEnv("DEV_MODE")

	app := &App{
		config: config,
		wsUpgrader: websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
			return true
		}},
		smRepo:         smRepo,
		uRepo:          uRepo,
		eventProcessor: eventprocessor.New(repos, bc, useCases),
		useCases:       useCases,
		wsApi:          api.NewWsApi(useCases, repos, registerer),
		events:         events,
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
	handleFunc("ping", pingHandler)
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
