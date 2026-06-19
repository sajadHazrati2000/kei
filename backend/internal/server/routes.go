package server

import (
	"net/http"

	"github.com/sajadHazrati2000/kei/backend/internal/availability"
	authpkg "github.com/sajadHazrati2000/kei/backend/internal/auth"
	"github.com/sajadHazrati2000/kei/backend/internal/middleware"
	"github.com/sajadHazrati2000/kei/backend/internal/realtime"
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

	// Availability — per-user
	avh := availability.NewHandler(s.availSvc, s.hub)
	s.mux.Handle("GET /api/v1/availability/{user_id}", requireAuth(http.HandlerFunc(avh.HandleGetSlots)))
	s.mux.Handle("PUT /api/v1/availability/{user_id}", requireAuth(http.HandlerFunc(avh.HandleSetSlots)))
	s.mux.Handle("GET /api/v1/availability/{user_id}/recurring", requireAuth(http.HandlerFunc(avh.HandleGetTemplates)))
	s.mux.Handle("PUT /api/v1/availability/{user_id}/recurring", requireAuth(http.HandlerFunc(avh.HandleSetTemplates)))
	s.mux.Handle("POST /api/v1/availability/{user_id}/import", requireAuth(http.HandlerFunc(avh.HandleImport)))

	// Team
	s.mux.Handle("GET /api/v1/team/availability", requireAuth(http.HandlerFunc(avh.HandleTeamAvailability)))
	s.mux.Handle("GET /api/v1/team/overlap", requireAuth(http.HandlerFunc(avh.HandleTeamOverlap)))

	// WebSocket — auth via ?token= query param (cookies unavailable during WS handshake)
	wsh := realtime.NewWSHandler(s.hub, s.cfg.JWTSecret, s.cfg.CORSOrigin, s.teamSnapshotFn)
	s.mux.Handle("GET /ws/availability", wsh)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
