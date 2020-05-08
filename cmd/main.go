package main

import (
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
	app, err = server.NewApp(appConfig)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}

	if err := app.Run(); err != nil {
		log.Fatalf("%s", err.Error())
	}
}
