package main

import (
	"log"
	"service-discovery/internal/configs"
	"service-discovery/internal/server"
)

func main() {
	cfg, err := configs.NewConfiguration()
	if err != nil {
		log.Panic(err)
	}

	server.NewServer(cfg.Server.Host, cfg.Server.Port).Run()
}
