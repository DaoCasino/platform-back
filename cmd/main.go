package main

import (
	"context"
	"log"
	"os"
	"platform-backend/config"
	"platform-backend/server"
)

func main() {
	confPath, isSet := os.LookupEnv("CONFIG_PATH")
	if !isSet {
		confPath = "config.json"
	}

	appConfig, err := config.Read(confPath)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}

	var app *server.App
	ctx := context.Background()
	app, err = server.NewApp(appConfig, ctx)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}

	if err := app.Run(ctx); err != nil {
		log.Fatalf("%s", err.Error())
	}
}
