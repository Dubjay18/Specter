package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Dubjay/specter/internal/config"
	"github.com/Dubjay/specter/internal/divergence"
	"github.com/Dubjay/specter/internal/proxy"
	"github.com/Dubjay/specter/internal/ring"
	"github.com/Dubjay/specter/internal/store"
)

func main() {
	configPath := "internal/config/specter.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	if _, err := os.Stat(configPath); err != nil {
		legacyCandidate := filepath.Join("internal", "config", filepath.Base(configPath))
		if _, legacyErr := os.Stat(legacyCandidate); legacyErr == nil {
			configPath = legacyCandidate
		}
	}

	cfg, err := config.Load(configPath)
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
	r := ring.NewRing(150)
	// eventDelegate := ring.NewEventDelegate(r)
	ml, err := ring.StartMembership(cfg.Cluster.NodeName, cfg.Cluster.BindAddr, cfg.Cluster.Peers, r)
	if err != nil {
		log.Fatalf("specter: failed to start membership: %v", err)
	}
	r.AddNode(cfg.Cluster.NodeName)
	defer ml.Leave(5 * time.Second)
	p := proxy.New(
		cfg.Specter.LiveTarget,
		cfg.Specter.ShadowTarget,
		cfg.Cluster.NodeName,
		cfg.Specter.RoutingKey,
		r,
		engine,
	)

	log.Printf("specter: listening on %s", cfg.Specter.Listen)
	log.Printf("specter: live    → %s", cfg.Specter.LiveTarget)
	log.Printf("specter: shadow  → %s", cfg.Specter.ShadowTarget)

	if err := http.ListenAndServe(cfg.Specter.Listen, p); err != nil {
		log.Fatalf("specter: server error: %v", err)
	}
}
