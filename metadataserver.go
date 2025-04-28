package metadataserver

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"path"
	"strconv"
	"time"
)

type Server struct {
	logger *slog.Logger
	config *Configuration
	server *http.Server
}

type Option func(*Server)

func WithConfiguration(c *Configuration) Option {
	return func(s *Server) {
		s.config = c
	}
}

func WithLogger(l *slog.Logger) Option {
	return func(s *Server) {
		s.logger = l
	}
}

func WithConfigFile(path string) Option {
	return func(s *Server) {
		c, err := NewConfigFromFile(path)
		if err != nil {
			if s.logger != nil {
				s.logger.Error("Failed to load config from file '%s': %s", path, err.Error())
			}
			return
		}
		s.config = c
	}
}

func NewServer(opts ...Option) (*Server, error) {
	s := &Server{config: NewConfiguration(DefaultConfigurationHandlers)}
	for _, opt := range opts {
		opt(s)
	}
	if s.logger == nil {
		s.logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	if s.config.Endpoint[0] != '/' {
		s.config.Endpoint = "/" + s.config.Endpoint
	}
	mux := http.NewServeMux()
	mux.HandleFunc(s.config.Endpoint, func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "ok")
	})
	for k, v := range s.config.Handlers {
		urlPath := path.Join(s.config.Endpoint, k)
		mux.HandleFunc(urlPath, func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, v())
		})
	}
	httpServer := &http.Server{
		Addr:    net.JoinHostPort(s.config.Address, strconv.Itoa(s.config.Port)),
		Handler: mux,
	}
	s.server = httpServer
	s.logger.DebugContext(context.Background(), "server is created", slog.Any("configuration", s.config))
	return s, nil
}

func (s *Server) Configuration() Configuration {
	return *s.config
}

func (s *Server) HttpHandler() http.Handler {
	if s.server == nil {
		return nil
	}
	return s.server.Handler
}

func (s *Server) Start() {
	s.logger.DebugContext(context.Background(), "starting metadata server", slog.Any("configuration", s.config))

	go func() {
		var srv = s.server
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.ErrorContext(context.Background(), "error listening and serving", slog.String("error", err.Error()))
		}
	}()
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.DebugContext(ctx, "stopping metadata server", slog.Any("configuration", s.config))
	shutdownCtx := context.Background()
	shutdownCtx, cancel := context.WithTimeout(shutdownCtx, time.Duration(s.config.ShutdownTimeout)*time.Second)
	defer cancel()
	return s.server.Shutdown(shutdownCtx)
}
