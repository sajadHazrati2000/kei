package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sajadHazrati2000/kei/backend/internal/auth"
	"github.com/sajadHazrati2000/kei/backend/internal/availability"
	"github.com/sajadHazrati2000/kei/backend/internal/config"
	"github.com/sajadHazrati2000/kei/backend/internal/realtime"
	"github.com/sajadHazrati2000/kei/backend/internal/settings"
	"github.com/sajadHazrati2000/kei/backend/internal/user"
)

type Server struct {
	cfg         *config.Config
	pool        *pgxpool.Pool
	mux         *http.ServeMux
	authSvc     *auth.Service
	userSvc     *user.Service
	availSvc    *availability.Service
	settingsSvc *settings.Service
	hub         *realtime.HubRegistry
}

func New(cfg *config.Config, pool *pgxpool.Pool) *Server {
	authRepo := auth.NewRepository(pool)
	authSvc := auth.NewService(authRepo, cfg.JWTSecret, cfg.JWTRefreshSecret)

	userRepo := user.NewRepository(pool)
	userSvc := user.NewService(userRepo)

	availRepo := availability.NewRepository(pool)
	availSvc := availability.NewService(availRepo)

	settingsRepo := settings.NewRepository(pool)
	settingsSvc := settings.NewService(settingsRepo)

	hub := realtime.NewHubRegistry()

	s := &Server{
		cfg:         cfg,
		pool:        pool,
		mux:         http.NewServeMux(),
		authSvc:     authSvc,
		userSvc:     userSvc,
		availSvc:    availSvc,
		settingsSvc: settingsSvc,
		hub:         hub,
	}
	s.registerRoutes()
	return s
}

// teamSnapshotFn fetches the current team availability and returns it as a
// pre-serialised JSON WebSocket message. Injected into the WS handler to keep
// the realtime package free of availability types.
func (s *Server) teamSnapshotFn(ctx context.Context, orgID string) ([]byte, error) {
	now := time.Now().UTC()
	from := now.Truncate(24 * time.Hour).AddDate(0, 0, -1)
	to := from.AddDate(0, 0, 15)
	team, err := s.availSvc.GetTeamAvailability(ctx, orgID, from, to)
	if err != nil {
		return nil, fmt.Errorf("server.teamSnapshotFn: %w", err)
	}
	return json.Marshal(map[string]any{
		"type": "snapshot",
		"data": team,
	})
}

func (s *Server) Run() error {
	return http.ListenAndServe(fmt.Sprintf(":%s", s.cfg.Port), corsMiddleware(s.cfg.CORSOrigin, s.mux))
}
