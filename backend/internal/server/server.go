package server

import (
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sajadHazrati2000/kei/backend/internal/config"
)

type Server struct {
	cfg  *config.Config
	pool *pgxpool.Pool
	mux  *http.ServeMux
}

func New(cfg *config.Config, pool *pgxpool.Pool) *Server {
	s := &Server{
		cfg:  cfg,
		pool: pool,
		mux:  http.NewServeMux(),
	}
	s.registerRoutes()
	return s
}

func (s *Server) Run() error {
	return http.ListenAndServe(fmt.Sprintf(":%s", s.cfg.Port), corsMiddleware(s.cfg.CORSOrigin, s.mux))
}
