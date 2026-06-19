package server

import (
	"net/http"

	authpkg "github.com/sajadHazrati2000/kei/backend/internal/auth"
	"github.com/sajadHazrati2000/kei/backend/internal/middleware"
	"github.com/sajadHazrati2000/kei/backend/internal/user"
)

func (s *Server) registerRoutes() {
	requireAuth := middleware.RequireAuth(s.cfg.JWTSecret)
	adminOnly := middleware.RequireRole("admin")

	s.mux.HandleFunc("GET /api/v1/health", s.handleHealth)

	// Auth — public
	ah := authpkg.NewHandler(s.authSvc, s.cfg.Env)
	s.mux.HandleFunc("POST /api/v1/auth/setup", ah.HandleSetup)
	s.mux.HandleFunc("POST /api/v1/auth/login", ah.HandleLogin)
	s.mux.HandleFunc("POST /api/v1/auth/refresh", ah.HandleRefresh)
	s.mux.HandleFunc("DELETE /api/v1/auth/logout", ah.HandleLogout)

	// Users
	uh := user.NewHandler(s.userSvc)
	s.mux.Handle("GET /api/v1/users", requireAuth(http.HandlerFunc(uh.HandleList)))
	s.mux.Handle("POST /api/v1/users/invite", requireAuth(adminOnly(http.HandlerFunc(uh.HandleInvite))))
	s.mux.Handle("GET /api/v1/users/{id}", requireAuth(http.HandlerFunc(uh.HandleGet)))
	s.mux.Handle("PUT /api/v1/users/{id}", requireAuth(http.HandlerFunc(uh.HandleUpdateProfile)))
	s.mux.Handle("PUT /api/v1/users/{id}/role", requireAuth(adminOnly(http.HandlerFunc(uh.HandleUpdateRole))))
	s.mux.Handle("DELETE /api/v1/users/{id}", requireAuth(adminOnly(http.HandlerFunc(uh.HandleDeactivate))))
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
