package main

import (
	"github.com/bradenrayhorn/switchboard-chat/config"
	"github.com/bradenrayhorn/switchboard-chat/database"
	"github.com/bradenrayhorn/switchboard-chat/grpc"
	"github.com/bradenrayhorn/switchboard-chat/hub"
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

	log.Printf("starting servers...")
	startServers()
}

func startServers() {
	// start gRPC
	log.Print("starting grpc client...")
	grpcClient := grpc.NewClient()

	// start chat hub
	log.Print("starting hub...")
	chatHub := hub.NewHub(&grpcClient)
	go chatHub.Start()

	// start gin router
	log.Print("starting http server...")
	r := routing.MakeRouter(&chatHub)

	err := r.Run()

	if err != nil {
		panic(err)
	}
}
