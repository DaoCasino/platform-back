package server

import (
	"context"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"os/signal"
	"platform-backend/auth"
	authPgRepo "platform-backend/auth/repository/postgres"
	authUC "platform-backend/auth/usecase"
	"platform-backend/config"
	"platform-backend/db"
	"platform-backend/logger"
	"platform-backend/models"
	"platform-backend/server/sessions"
	"time"
)

type App struct {
	httpServer     *http.Server
	config         *config.Config
	wsUpgrader     websocket.Upgrader
	sessionManager *sessions.SessionManager

	authUC auth.UseCase
}

func wsClientHandler(app *App, w http.ResponseWriter, r *http.Request) {
	log.Debug().Msgf("New connect request")

	token, err := r.Cookie("token")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Debug().Msgf("Token cookie not found")
		return
	}

	user, err := app.authUC.ParseToken(context.Background(), token.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		log.Debug().Msgf("Invalid auth token")
		return
	}

	c, err := app.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Err(err)
		return
	}

	log.Info().Msgf("Client with ip %q connected", c.RemoteAddr())

	app.sessionManager.NewConnection(user.AccountName, c)
}

func authHandler(app *App, w http.ResponseWriter, r *http.Request) {
	log.Debug().Msgf("New auth request")

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Debug().Msgf("Http body parse error, %s", err.Error())
		return
	}

	signedToken, err := app.authUC.SignUp(context.Background(), &user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	authCookie := &http.Cookie{Name: "token", Value: signedToken, HttpOnly: true}
	http.SetCookie(w, authCookie)
	w.WriteHeader(http.StatusOK)
}

func NewApp(config *config.Config) (*App, error) {
	logger.InitLogger(config.LogLevel)

	err := db.InitDB(context.Background(), &config.DbConfig)
	if err != nil {
		log.Fatal().Msgf("Database init error, %s", err.Error())
		return nil, err
	}

	app := &App{
		config: config,
		wsUpgrader: websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
			return true
		}},
		sessionManager: sessions.NewSessionManager(),
		authUC: authUC.NewAuthUseCase(
			authPgRepo.NewUserPostgresRepo(db.DbPool),
			[]byte(config.AuthConfig.JwtSecret),
		),
	}

	wsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wsClientHandler(app, w, r)
	})

	authHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHandler(app, w, r)
	})

	http.Handle("/connect", wsHandler)
	http.HandleFunc("/auth", authHandler)

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
