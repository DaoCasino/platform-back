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

	app := server.NewApp(appConfig)

	if err := app.Run(appConfig.Port); err != nil {
		log.Fatalf("%s", err.Error())
	}
}
