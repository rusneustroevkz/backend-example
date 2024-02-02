package http

import (
	"context"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net"
	"net/http"
)

type Server struct {
	log *zap.Logger
	srv *http.Server
}

func NewHTTPServer(mux *chi.Mux, log *zap.Logger) *Server {
	srv := &http.Server{Addr: ":8080", Handler: mux}

	return &Server{
		log: log,
		srv: srv,
	}
}

func (s *Server) Start(_ context.Context) error {
	listener, err := net.Listen("tcp", s.srv.Addr)
	if err != nil {
		return err
	}
	s.log.Info("Starting HTTP server", zap.String("addr", s.srv.Addr))
	go func() {
		err := s.srv.Serve(listener)
		s.log.Fatal("cannot start server", zap.Error(err))
	}()
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}