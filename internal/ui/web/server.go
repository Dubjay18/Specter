package web

import (
	"encoding/json"
	"net/http"

	"github.com/Dubjay/specter/internal/divergence"
)

type StatsProvider interface {
	StatsSnapshot() divergence.StatsSnapshot
}

type Server struct {
	statsProvider StatsProvider
}

func NewServer(statsProvider StatsProvider) *Server {
	return &Server{statsProvider: statsProvider}
}

func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/stats", s.handleStats)
}

func (s *Server) handleStats(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(s.statsProvider.StatsSnapshot()); err != nil {
		http.Error(w, "failed to encode stats", http.StatusInternalServerError)
		return
	}
}
