package server

import (
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sajadHazrati2000/kei/backend/internal/auth"
	"github.com/sajadHazrati2000/kei/backend/internal/availability"
	"github.com/sajadHazrati2000/kei/backend/internal/config"
	"github.com/sajadHazrati2000/kei/backend/internal/user"
)

type Server struct {
	cfg      *config.Config
	pool     *pgxpool.Pool
	mux      *http.ServeMux
	authSvc  *auth.Service
	userSvc  *user.Service
	availSvc *availability.Service
}

func New(cfg *config.Config, pool *pgxpool.Pool) *Server {
	authRepo := auth.NewRepository(pool)
	authSvc := auth.NewService(authRepo, cfg.JWTSecret, cfg.JWTRefreshSecret)

	userRepo := user.NewRepository(pool)
	userSvc := user.NewService(userRepo)

	availRepo := availability.NewRepository(pool)
	availSvc := availability.NewService(availRepo)

	s := &Server{
		cfg:      cfg,
		pool:     pool,
		mux:      http.NewServeMux(),
		authSvc:  authSvc,
		userSvc:  userSvc,
		availSvc: availSvc,
	}
	s.registerRoutes()
	return s
}

func (s *Server) Run() error {
	return http.ListenAndServe(fmt.Sprintf(":%s", s.cfg.Port), corsMiddleware(s.cfg.CORSOrigin, s.mux))
}
