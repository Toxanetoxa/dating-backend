package main

import (
	"log"

	"github.com/toxanetoxa/dating-backend/config"
	"github.com/toxanetoxa/dating-backend/internal/app"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	err = app.Run(cfg)
	if err != nil {
		log.Fatalf("app error: %v", err)
	}
}
