package main

import (
	"log"

	"github.com/keshvan/forum-service-sstu-forum/config"
	"github.com/keshvan/forum-service-sstu-forum/internal/app"
)

// @title Forum Service API
// @version 1.0
// @description API for forum service
// @host localhost:3000
// @BasePath /
func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}
	app.Run(cfg)
}
