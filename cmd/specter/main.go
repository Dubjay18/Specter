package main

import (
	"log"
	"net/http"
	"github.com/Dubjay/specter/internal/config"
	"github.com/Dubjay/specter/internal/proxy"
)


func main() {
cfg, err := config.Load("config/specter.yaml")
	if err != nil {
		log.Fatalf("specter: failed to load config: %v", err)
	}

	p := proxy.New(cfg.Specter.LiveTarget, cfg.Specter.ShadowTarget)

	log.Printf("specter: listening on %s", cfg.Specter.Listen)
	log.Printf("specter: live    → %s", cfg.Specter.LiveTarget)
	log.Printf("specter: shadow  → %s", cfg.Specter.ShadowTarget)

	if err := http.ListenAndServe(cfg.Specter.Listen, p); err != nil {
		log.Fatalf("specter: server error: %v", err)
	}
}