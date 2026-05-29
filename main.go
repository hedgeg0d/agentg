package main

import (
	"flag"
	"log"

	"github.com/hedgeg0d/agentg/internal/bot"
	"github.com/hedgeg0d/agentg/internal/config"
	"github.com/hedgeg0d/agentg/internal/store"
)

func main() {
	cfgPath := flag.String("config", "config.json", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	st, err := store.Open(cfg.DataDir)
	if err != nil {
		log.Fatalf("store: %v", err)
	}
	b, err := bot.New(cfg, st)
	if err != nil {
		log.Fatalf("bot: %v", err)
	}
	b.Run()
}
