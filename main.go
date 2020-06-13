package main

import (
	"github.com/bradenrayhorn/switchboard-chat/config"
	"github.com/bradenrayhorn/switchboard-chat/database"
	"github.com/bradenrayhorn/switchboard-chat/routing"
	"log"
)

func main() {
	log.Printf("starting switchboard chat...")

	log.Printf("loading config...")
	config.LoadConfig()
	log.Printf("config loaded!")

	log.Printf("connecting to database...")
	database.Setup()
	log.Printf("database connected!")

	log.Printf("starting server...")
	startServer()
}

func startServer() {
	// start chat hub

	// start gin router
	r := routing.MakeRouter()

	err := r.Run()

	if err != nil {
		panic(err)
	}
}
