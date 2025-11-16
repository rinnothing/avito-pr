package main

import (
	"log"

	"github.com/rinnothing/avito-pr/config"
	"github.com/rinnothing/avito-pr/internal/api"
	"github.com/rinnothing/avito-pr/pkg/logger"
)

func main() {
	cfg, err := config.New("config/prod.yaml")
	if err != nil {
		log.Fatalf("can't initialize config: %s", err.Error())
	}

	lg, err := logger.ConstructLogger(cfg.Logger.Level, cfg.Logger.Filepath)
	if err != nil {
		log.Fatalf("can't initialize logger: %s", err.Error())
	}
	defer lg.Sync()

	s := api.Server{}
	s.Run(lg, cfg)
}
