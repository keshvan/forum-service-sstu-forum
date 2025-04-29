package main

import (
	"log"

	"github.com/keshvan/forum-service-sstu-forum/config"
	"github.com/keshvan/forum-service-sstu-forum/internal/app"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}
	app.Run(cfg)
}
