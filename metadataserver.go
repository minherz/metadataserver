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

var (
	// ErrServerAlreadyStarted indicates that the server is already running.
	ErrServerAlreadyStarted error = errors.New("server has been already started")
	// ErrServerIsNotRunning indicates that the server has not been started.
	ErrServerIsNotRunning error = errors.New("server is not running")
)

// Server implements an instance of the metadata server.
// It registers handlers based on the paths that are defined in the configuration.
// The handlers can return either a literal or a value of the environment variable.
type Server struct {
	logger *slog.Logger
	config *Configuration
	server *http.Server
	status chan error
}

// Option allows to set up an instance of Server at creation time.
type Option func(*Server)

// WithAddress sets a new server with an IP address at which server accepts requests.
//
// Mind the order of options when use with [WithConfiguration] and [WithConfigFile].
func WithAddress(address string) Option {
	return func(s *Server) {
		if s.config == nil {
			s.config = NewConfiguration(DefaultConfigurationHandlers)
		}
		s.config.Address = address
	}
}

// WithConfigFile sets a new server with [Configuration] that is read from JSON file.
//
// Mind the order of options when use with [WithConfiguration], [WithAddress], [WithPort] and [WithHandlers].
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

// WithConfiguration sets a new server with [Configuration].
//
// Mind the order of options when use with [WithConfigFile], [WithAddress], [WithPort] and [WithHandlers].
func WithConfiguration(c *Configuration) Option {
	return func(s *Server) {
		s.config = c
	}
}

// WithEndpoint sets a new default endpoint path.
//
// Mind the order of options when use with [WithConfiguration] and [WithConfigFile].
func WithEndpoint(path string) Option {
	return func(s *Server) {
		if s.config == nil {
			s.config = NewConfiguration(DefaultConfigurationHandlers)
		}
		s.config.Endpoint = path
	}
}

// WithHandlers sets a new server with a set of metadata handlers.
//
// Mind the order of options when use with [WithConfiguration] and [WithConfigFile].
func WithHandlers(handlers map[string]Metadata) Option {
	return func(s *Server) {
		if s.config == nil {
			s.config = NewConfiguration(DefaultConfigurationHandlers)
		}
		s.config.Handlers = handlers
	}
}

// WithPort sets a new server with a port number at which server accepts requests.
//
// Mind the order of options when use with [WithConfiguration] and [WithConfigFile].
func WithPort(port int) Option {
	return func(s *Server) {
		if s.config == nil {
			s.config = NewConfiguration(DefaultConfigurationHandlers)
		}
		s.config.Port = port
	}
}

// WithLogger sets a new server with an instance of [slog.Logger].
func WithLogger(l *slog.Logger) Option {
	return func(s *Server) {
		s.logger = l
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

// Start launches the server to server configured metadata handlers.
//
// It returns ErrServerHasBeenStarted if the server has already been started.
// Otherwise it return an error if failed to start serving on the configured address.
func (s *Server) Start(ctx context.Context) error {
	if s.status != nil {
		return ErrServerAlreadyStarted
	}
	s.logger.DebugContext(ctx, "starting metadata server", slog.Any("configuration", s.config))
	s.status = make(chan error)
	go func() {
		err := s.server.ListenAndServe()
		s.status <- err
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.ErrorContext(ctx, "error listening and serving", slog.String("error", err.Error()))
		}
	}()
	select {
	case err := <-s.status:
		s.status = nil
		return err
	case <-time.After(100 * time.Millisecond):
	}
	return nil
}

// Stop shuts down the running server.
//
// It returns ErrServerIsNotRunning if the server was not started.
// Otherwise it return an error if failed to stop the running service.
func (s *Server) Stop(ctx context.Context) error {
	if s.status == nil {
		return ErrServerIsNotRunning
	}
	s.logger.DebugContext(ctx, "stopping metadata server", slog.Any("configuration", s.config))
	s.status = nil
	shutdownCtx := context.Background()
	shutdownCtx, cancel := context.WithTimeout(shutdownCtx, time.Duration(s.config.ShutdownTimeout)*time.Second)
	defer cancel()
	return s.server.Shutdown(shutdownCtx)
}
