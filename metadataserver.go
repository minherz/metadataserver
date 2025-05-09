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

// Server implements an instance of the metadata server.
// It registers handlers based on the paths that are defined in the configuration.
// The handlers can return either a literal or a value of the environment variable.
type Server struct {
	logger *slog.Logger
	config *Configuration
	server *http.Server
}

// Option allows to set up an instance of Server at creation time.
type Option func(*Server)

// WithConfiguration creates an Option that sets up a new configuration for the server.
// DO NOT use it with `WithConfigFile`, `WithAddress` or `WithPort` option.
func WithConfiguration(c *Configuration) Option {
	return func(s *Server) {
		s.config = c
	}
}

// WithAddress creates an Option that sets up a serving address for the server.
// DO NOT use it with `WithConfigFile` or `WithConfiguration` option.
func WithAddress(address string) Option {
	return func(s *Server) {
		if s.config == nil {
			s.config = NewConfiguration(DefaultConfigurationHandlers)
		}
		s.config.Address = address
	}
}

// WithAddress creates an Option that sets up a listening port for the server.
// DO NOT use it with `WithConfiguration`, `WithAddress` or `WithPort` option.
func WithPort(port int) Option {
	return func(s *Server) {
		if s.config == nil {
			s.config = NewConfiguration(DefaultConfigurationHandlers)
		}
		s.config.Port = port
	}
}

// WithLogger creates an Option that sets up `slog.Logger` for the server.
func WithLogger(l *slog.Logger) Option {
	return func(s *Server) {
		s.logger = l
	}
}

// WithConfigFile creates an Option that loads a server configuration from a file.
// DO NOT use it with `WithConfiguration` option.
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

// New creates a new instance of the server.
func New(opts ...Option) (*Server, error) {
	s := &Server{}
	for _, opt := range opts {
		opt(s)
	}
	if s.config == nil {
		s.config = NewConfiguration(DefaultConfigurationHandlers)
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
		mux.HandleFunc(urlPath, func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			data := v()
			s.logger.DebugContext(ctx, "metadata handler is called",
				slog.String("handler", r.URL.Path), slog.String("response", data))
			fmt.Fprint(w, data)
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

// Configuration returns a copy of the server's configuration
func (s *Server) Configuration() Configuration {
	return *s.config
}

// HttpHandler returns collection of HTTP handlers
func (s *Server) HttpHandler() http.Handler {
	if s.server == nil {
		return nil
	}
	return s.server.Handler
}

// Start launches the server that will listen at the configured address and port
// and will serve the metadata.
func (s *Server) Start(ctx context.Context) error {
	s.logger.DebugContext(ctx, "starting metadata server", slog.Any("configuration", s.config))
	ch := make(chan error)
	go func() {
		var srv = s.server
		err := srv.ListenAndServe()
		ch <- err
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.ErrorContext(ctx, "error listening and serving", slog.String("error", err.Error()))
		}
	}()
	select {
	case err := <-ch:
		return err
	case <-time.After(100 * time.Millisecond):
	}
	return nil
}

// Stop shuts down the running server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.DebugContext(ctx, "stopping metadata server", slog.Any("configuration", s.config))
	shutdownCtx := context.Background()
	shutdownCtx, cancel := context.WithTimeout(shutdownCtx, time.Duration(s.config.ShutdownTimeout)*time.Second)
	defer cancel()
	return s.server.Shutdown(shutdownCtx)
}
