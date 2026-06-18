package server

import (
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sajadHazrati2000/kei/backend/internal/auth"
	"github.com/sajadHazrati2000/kei/backend/internal/config"
)

type Server struct {
	cfg     *config.Config
	pool    *pgxpool.Pool
	mux     *http.ServeMux
	authSvc *auth.Service
}

func New(cfg *config.Config, pool *pgxpool.Pool) *Server {
	authRepo := auth.NewRepository(pool)
	authSvc := auth.NewService(authRepo, cfg.JWTSecret, cfg.JWTRefreshSecret)

	s := &Server{
		cfg:     cfg,
		pool:    pool,
		mux:     http.NewServeMux(),
		authSvc: authSvc,
	}
	s.registerRoutes()
	return s
}

func (s *Server) Run() error {
	return http.ListenAndServe(fmt.Sprintf(":%s", s.cfg.Port), corsMiddleware(s.cfg.CORSOrigin, s.mux))
}
