package main

import (
	"log"
	"platform-backend/config"
	"platform-backend/server"
)

func main() {
	appConfig, err := config.FromFile("config.json")
	if err != nil {
		log.Fatalf("%s", err.Error())
	}

	var app *server.App
	app, err = server.NewApp(appConfig)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}

	if err := app.Run(appConfig.Port); err != nil {
		log.Fatalf("%s", err.Error())
	}
}
