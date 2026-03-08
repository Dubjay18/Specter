package main

import (
	"log"
	"net/http"

	"github.com/Dubjay/specter/internal/config"
	"github.com/Dubjay/specter/internal/divergence"
	"github.com/Dubjay/specter/internal/proxy"
	"github.com/Dubjay/specter/internal/store"
)

func main() {
	cfg, err := config.Load("config/specter.yaml")
	if err != nil {
		log.Fatalf("specter: failed to load config: %v", err)
	}
	//  open the store on startup and inject it into the divergence engine
	eventStore, err := store.NewStore(cfg.Store.BadgerPath)
	if err != nil {
		log.Fatalf("specter: failed to initialize store: %v", err)
	}
	defer eventStore.Close()

	engine := divergence.NewEngine(eventStore)

	p := proxy.New(cfg.Specter.LiveTarget, cfg.Specter.ShadowTarget, engine)

	log.Printf("specter: listening on %s", cfg.Specter.Listen)
	log.Printf("specter: live    → %s", cfg.Specter.LiveTarget)
	log.Printf("specter: shadow  → %s", cfg.Specter.ShadowTarget)

	if err := http.ListenAndServe(cfg.Specter.Listen, p); err != nil {
		log.Fatalf("specter: server error: %v", err)
	}
}
