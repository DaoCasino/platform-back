package server

import (
	"context"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"os/signal"
	authPgRepo "platform-backend/auth/repository/postgres"
	authUC "platform-backend/auth/usecase"
	casinoPgRepo "platform-backend/casino/repository/postgres"
	casinoUC "platform-backend/casino/usecase"
	"platform-backend/config"
	"platform-backend/db"
	"platform-backend/logger"
	"platform-backend/models"
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
	httpServer *http.Server
	config     *config.Config
	wsUpgrader websocket.Upgrader
	wsApi      *api.WsApi

	smRepo   session_manager.Repository
	useCases *usecases.UseCases
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
	w.Write(response)
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
	w.Write(response)
}

func NewApp(config *config.Config) (*App, error) {
	logger.InitLogger(config.LogLevel)

	err := db.InitDB(context.Background(), &config.DbConfig)
	if err != nil {
		log.Fatal().Msgf("Database init error, %s", err.Error())
		return nil, err
	}

	smRepo := smLocalRepo.NewLocalRepository()

	useCases := usecases.NewUseCases(
		authUC.NewAuthUseCase(
			authPgRepo.NewUserPostgresRepo(db.DbPool),
			smRepo,
			[]byte(config.AuthConfig.JwtSecret),
			config.AuthConfig.AccessTokenTTL,
			config.AuthConfig.RefreshTokenTTL,
		),
		casinoUC.NewCasinoUseCase(
			casinoPgRepo.NewCasinoPostgresRepo(db.DbPool),
		),
	)

	app := &App{
		config: config,
		wsUpgrader: websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
			return true
		}},
		smRepo:   smRepo,
		useCases: useCases,
		wsApi:    api.NewWsApi(useCases),
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

	http.Handle("/connect", wsHandler)
	http.HandleFunc("/auth", authHandler)
	http.HandleFunc("/refresh_token", refreshTokensHandler)

	log.Info().Msg("App created")

	return app, nil
}

func (a *App) Run(port string) error {
	log.Info().Msgf("Server is listening on %s port", a.config.Port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		return err
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Interrupt)

	<-quit

	ctx, shutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdown()

	return a.httpServer.Shutdown(ctx)
}
