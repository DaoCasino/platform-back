package server

import (
	"context"
	"encoding/json"
	"github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"os/signal"
	"platform-backend/config"
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
}

func wsClientHandler(app *App, w http.ResponseWriter, r *http.Request) {
	log.Debug().Msgf("New connect request")

	user := r.Context().Value("user")
	jwtData := user.(*jwt.Token).Claims.(jwt.MapClaims)

	c, err := app.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Err(err)
		return
	}

	log.Info().Msgf("Client with ip %q connected", c.RemoteAddr())

	app.sessionManager.NewConnection(jwtData["account_name"].(string), c)
}

func authHandler(app *App, w http.ResponseWriter, r *http.Request) {
	log.Debug().Msgf("New auth request")

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Debug().Msgf("Http body parse error, %s", err.Error())
		return
	}

	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = jwt.MapClaims{"account_name": user.AccountName}

	signed, err := token.SignedString([]byte(app.config.AuthConfig.JwtSecret))
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		log.Debug().Msgf("JWT token sign error, %s", err.Error())
		return
	}

	hasUser, err := models.HasUser(context.Background(), user.AccountName)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		log.Debug().Msgf("User existing check error, %s", err.Error())
		return
	}

	if !hasUser {
		err = models.AddUser(context.Background(), &user)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			log.Debug().Msgf("User create error, %s", err.Error())
			return
		}
	}

	authCookie := &http.Cookie{Name: "token", Value: signed, HttpOnly: true}
	http.SetCookie(w, authCookie)
	w.WriteHeader(http.StatusOK)
}

func NewApp(config *config.Config) (*App, error) {
	logger.InitLogger(config.LogLevel)

	err := models.InitDB(context.Background(), &config.DbConfig)
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
	}

	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (i interface{}, err error) {
			return []byte(config.AuthConfig.JwtSecret), nil
		},
		SigningMethod: jwt.SigningMethodHS256,
		Extractor: func(r *http.Request) (string, error) {
			token, err := r.Cookie("token")
			if err != nil {
				return "", err
			}
			return token.Value, nil
		},
	})

	wsHandler := jwtMiddleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wsClientHandler(app, w, r)
	}))

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
