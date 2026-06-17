package server

import (
	"net/http"
)

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("GET /api/v1/health", s.handleHealth)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
