package server

import (
	"context"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"os/signal"
	"platform-backend/auth"
	"platform-backend/auth/repository/localstorage"
	"platform-backend/config"
	"platform-backend/logger"
	"time"

	authusecase "platform-backend/auth/usecase"
)

type App struct {
	httpServer *http.Server
	config     *config.Config
	upgrader   websocket.Upgrader

	authUC auth.UseCase
}

func wsMessageHandler(app *App, ws *websocket.Conn) {
	defer ws.Close()
	ws.SetReadLimit(512)
	for {
		messageType, message, err := ws.ReadMessage()
		if err != nil {
			// There is error or client is disconnected
			log.Info().Msgf("Client with ip %q disconnected", ws.RemoteAddr())
			break
		}
		log.Info().Msgf("Type: %d, message: %s", messageType, message)

		if string(message) == "auth" {
			app.authUC.SignIn(context.Background(), "petya", "qwerty")
		}

		//err = ws.WriteMessage(websocket.TextMessage, []byte("Hello, my client!"))
		//if err != nil {
		//	log.Info().Msgf("Client with ip %q disconnected", ws.RemoteAddr())
		//	break
		//}
	}
}

func wsClientHandler(app *App, w http.ResponseWriter, r *http.Request) {
	c, err := app.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Err(err)
		return
	}
	log.Info().Msgf("Client with ip %q connected", c.RemoteAddr())
	go wsMessageHandler(app, c)
}

func NewApp(config *config.Config) *App {
	logger.InitLogger(config.LogLevel)

	//db := initDB()

	//userRepo := authmongo.NewUserRepository(db, viper.GetString("mongo.user_collection"))

	userRepo := localstorage.NewUserLocalStorage()
	app := &App{
		config: config,
		upgrader: websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
			return true
		}},
		authUC: authusecase.NewAuthUseCase(
			userRepo,
			"test_hash_salt",
		),
	}

	http.HandleFunc("/connect", func(w http.ResponseWriter, r *http.Request) {
		wsClientHandler(app, w, r)
	})

	log.Info().Msg("App created")

	return app
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

func initDB() {
	// Init db here
	// Below is example for postgres

	//client, err := mongo.NewClient(options.Client().ApplyURI(viper.GetString("mongo.uri")))
	//if err != nil {
	//	log.Fatalf("Error occured while establishing connection to mongoDB")
	//}
	//
	//ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	//defer cancel()
	//
	//err = client.Connect(ctx)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//err = client.Ping(context.Background(), nil)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//return client.Database(viper.GetString("mongo.name"))
}
