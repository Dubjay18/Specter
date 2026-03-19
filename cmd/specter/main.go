package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Dubjay/specter/internal/config"
	"github.com/Dubjay/specter/internal/divergence"
	"github.com/Dubjay/specter/internal/proxy"
	"github.com/Dubjay/specter/internal/ring"
	"github.com/Dubjay/specter/internal/store"
	"github.com/Dubjay/specter/internal/ui/tui"
	"github.com/Dubjay/specter/internal/ui/web"
)

func main() {
	configPathFlag := flag.String("config", "internal/config/specter.yaml", "path to config file")
	uiMode := flag.String("ui", "proxy", "run mode: proxy or tui")
	flag.Parse()

	configPath := *configPathFlag
	if flag.NArg() > 0 {
		configPath = flag.Arg(0)
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

	if *uiMode == "tui" {
		statsURL := statsAPIURL(cfg.Specter.Listen)
		log.Printf("specter: launching tui dashboard against %s", statsURL)
		if err := tui.Run(statsURL); err != nil {
			log.Fatalf("specter: tui error: %v", err)
		}
		return
	}

	if *uiMode != "proxy" {
		log.Fatalf("specter: invalid --ui value %q (expected proxy or tui)", *uiMode)
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

	mux := http.NewServeMux()
	webServer := web.NewServer(engine)
	webServer.RegisterRoutes(mux)
	mux.Handle("/", p)

	if err := http.ListenAndServe(cfg.Specter.Listen, mux); err != nil {
		log.Fatalf("specter: server error: %v", err)
	}
}

func statsAPIURL(listenAddr string) string {
	listenAddr = strings.TrimSpace(listenAddr)
	if strings.HasPrefix(listenAddr, "http://") || strings.HasPrefix(listenAddr, "https://") {
		return strings.TrimRight(listenAddr, "/") + "/api/stats"
	}

	if strings.HasPrefix(listenAddr, ":") {
		return "http://127.0.0.1" + listenAddr + "/api/stats"
	}

	if strings.HasPrefix(listenAddr, "0.0.0.0:") {
		return "http://127.0.0.1:" + strings.TrimPrefix(listenAddr, "0.0.0.0:") + "/api/stats"
	}

	if listenAddr == "" {
		return "http://127.0.0.1:8080/api/stats"
	}

	return "http://" + listenAddr + "/api/stats"
}
