package server

import (
	"auto/internal/storage"
	"errors"
	"fmt"
	"github.com/pingcap/failpoint"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Server defines fields used in HTTP processing.
type Server struct {
	logger        *zap.Logger
	addr          string
	httpServer    *fasthttp.Server
	afterShutdown func() error
}

// New constructs a Server. See the various Options for available customizations.
func New(logger *zap.Logger, storage *storage.Storage, options ...Option) (Server, error) {
	if logger == nil {
		return Server{}, errors.New("no logger provided")
	}

	if storage == nil {
		return Server{}, errors.New("no storage provided")
	}

	config := &config{}

	h := handler{logger: logger, Storage: storage}
	m := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/api/shorten":
			h.saveURL(ctx)
		default:
			h.getURL(ctx)
		}
	}

	s := &fasthttp.Server{
		Handler:          m,
		DisableKeepalive: true,
		ReadTimeout:      5 * time.Second,
	}

	for _, o := range options {
		o.apply(config)
	}

	return Server{
		logger:        logger,
		addr:          config.addr,
		httpServer:    s,
		afterShutdown: storage.Close,
	}, nil
}

// Start calls ListenAndServe on fasthttp.Server instance inside Server struct
// and implements graceful shutdown via goroutine waiting for signals.
func (s *Server) Start() error {
	idleConnsClosed := make(chan struct{})

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
		<-sigint

		s.logger.Info("Shutting down HTTP server")

		err := s.httpServer.Shutdown()
		failpoint.Inject("shutdownErr", func() {
			err = errors.New("mock shutdown error")
		})
		if err != nil {
			s.logger.Error("srv.Shutdown: %v", zap.Error(err))
		}
		s.logger.Info("HTTP server is stopped")

		close(idleConnsClosed)
	}()

	s.logger.Info("Starting HTTP server", zap.String("address", s.addr))
	err := s.httpServer.ListenAndServe(s.addr)
	failpoint.Inject("listenAndServeErr", func() {
		err = errors.New("mock listen and serve error")
	})
	if err != nil {
		return fmt.Errorf("ListenAndServe error: %w", err)
	}

	<-idleConnsClosed

	if err := s.afterShutdown(); err != nil {
		return err
	}

	return nil
}
