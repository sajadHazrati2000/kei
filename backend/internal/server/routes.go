package server

import (
	"net/http"

	"github.com/sajadHazrati2000/kei/backend/internal/auth"
)

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("GET /api/v1/health", s.handleHealth)

	authHandler := auth.NewHandler(s.authSvc, s.cfg.Env)
	s.mux.HandleFunc("POST /api/v1/auth/setup", authHandler.HandleSetup)
	s.mux.HandleFunc("POST /api/v1/auth/login", authHandler.HandleLogin)
	s.mux.HandleFunc("POST /api/v1/auth/refresh", authHandler.HandleRefresh)
	s.mux.HandleFunc("DELETE /api/v1/auth/logout", authHandler.HandleLogout)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
