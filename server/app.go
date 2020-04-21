package server

import (
	"context"
	"encoding/json"
	eventlistener "github.com/DaoCasino/platform-action-monitor-client"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"os/signal"
	authPgRepo "platform-backend/auth/repository/postgres"
	authUC "platform-backend/auth/usecase"
	"platform-backend/blockchain"
	casinoBcRepo "platform-backend/casino/repository/blockchain"
	casinoUC "platform-backend/casino/usecase"
	"platform-backend/config"
	"platform-backend/db"
	gameSesssionPgRepo "platform-backend/game_sessions/repository/postgres"
	"platform-backend/eventprocessor"
	"platform-backend/logger"
	"platform-backend/models"
	"platform-backend/repositories"
	"platform-backend/server/api"
	"platform-backend/server/session_manager"
	smLocalRepo "platform-backend/server/session_manager/repository/localstorage"
	"platform-backend/usecases"
	"time"
)

type JsonResponse = map[string]interface{}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type App struct {
	httpHandler http.Handler
	config      *config.Config
	wsUpgrader  websocket.Upgrader
	wsApi       *api.WsApi

	smRepo   session_manager.Repository
	eventProcessor *eventprocessor.Processor
	useCases *usecases.UseCases
	events   chan *eventlistener.EventMessage
}

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
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Debug().Msgf("Http body parse error, %s", err.Error())
		return
	}

	log.Debug().Msgf("New auth request from %s", user.AccountName)

	refreshToken, accessToken, err := app.useCases.Auth.SignUp(context.Background(), &user)
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

func NewApp(config *config.Config) (*App, error) {
	logger.InitLogger(config.LogLevel)

	err := db.InitDB(context.Background(), &config.DbConfig)
	if err != nil {
		log.Fatal().Msgf("Database init error, %s", err.Error())
		return nil, err
	}

	bc, err := blockchain.Init(config.BlockchainConfig.NodeUrl, config.BlockchainConfig.SponsorUrl)
	if err != nil {
		log.Fatal().Msgf("Blockchain init error, %s", err.Error())
		return nil, err
	}

	smRepo := smLocalRepo.NewLocalRepository()
	repos := repositories.NewRepositories(
		casinoBcRepo.NewCasinoBlockchainRepo(bc, config.BlockchainConfig.Contracts.Platform),
		gameSesssionPgRepo.NewGameSessionsPostgresRepo(db.DbPool),
	)

	useCases := usecases.NewUseCases(
		authUC.NewAuthUseCase(
			authPgRepo.NewUserPostgresRepo(db.DbPool),
			smRepo,
			[]byte(config.AuthConfig.JwtSecret),
			config.AuthConfig.AccessTokenTTL,
			config.AuthConfig.RefreshTokenTTL,
		),
		casinoUC.NewCasinoUseCase(
			repos.Casino,
			repos.GameSession,
			bc,
			config.BlockchainConfig.Permissions.GameAction,
		),
	)

	events := make(chan *eventlistener.EventMessage)

	app := &App{
		config: config,
		wsUpgrader: websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
			return true
		}},
		smRepo:   smRepo,
		eventProcessor: eventprocessor.New(repos.GameSession),
		useCases: useCases,
		wsApi:    api.NewWsApi(useCases, repos),
		events:   events,
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

	mux := http.NewServeMux()

	mux.Handle("/connect", wsHandler)
	mux.HandleFunc("/auth", authHandler)
	mux.HandleFunc("/refresh_token", refreshTokensHandler)

	app.httpHandler = cors.Default().Handler(mux)

	log.Info().Msg("App created")

	return app, nil
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
	listener := eventlistener.NewEventListener(a.config.AmcConfig.Url, a.events)
	log.Info().Msgf("Connecting to the action monitor on %s", a.config.AmcConfig.Url)

	if err := listener.ListenAndServe(ctx); err != nil {
		log.Error().Msgf("Action monitor connect error: %v", err)
		return err
	}

	if a.config.LogLevel == "debug" {
		eventlistener.EnableDebugLogging()
	}

	if ok, err := listener.Subscribe(1, 0); err != nil || !ok {
		log.Error().Msgf("Action monitor subscribe error: %v", err)
		return err
	}

	log.Info().Msgf("Connected to the action monitor")

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
				// TODO: notify clients
				go a.eventProcessor.Process(ctx, event)
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
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, os.Interrupt)
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
