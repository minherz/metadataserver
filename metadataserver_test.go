package metadataserver_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/minherz/metadataserver"
)

func TestNewServer(t *testing.T) {
	testConfig := metadataserver.Configuration{
		Address:         "1.2.3.4",
		Endpoint:        "custom/endpoint",
		Port:            8080,
		ShutdownTimeout: 15,
		Handlers: map[string]metadataserver.Metadata{
			"entry1": func() string {
				return "one"
			},
		},
	}
	wantConfig := metadataserver.Configuration{
		Address:         "1.2.3.4",
		Endpoint:        "/custom/endpoint",
		Port:            8080,
		ShutdownTimeout: 15,
		Handlers: map[string]metadataserver.Metadata{
			"entry1": func() string {
				return "one"
			},
		},
	}
	tests := []struct {
		name  string
		input []metadataserver.Option
		want  metadataserver.Configuration
	}{
		{
			name:  "default",
			input: []metadataserver.Option{},
			want: metadataserver.Configuration{
				Address:         metadataserver.DefaultAddress,
				Endpoint:        metadataserver.DefaultEndpoint,
				Port:            metadataserver.DefaultPort,
				ShutdownTimeout: metadataserver.DefaultShutdownTimeout,
				Handlers:        metadataserver.DefaultConfigurationHandlers,
			},
		},
		{
			name: "opt_config_from_object",
			input: []metadataserver.Option{
				metadataserver.WithConfiguration(&testConfig),
			},
			want: wantConfig,
		},
		{
			name: "opt_config_from_file",
			input: []metadataserver.Option{
				metadataserver.WithConfigFile("test/fixtures/config_smoke_test.json"),
			},
			want: wantConfig,
		},
		{
			name: "opt_address_and_port",
			input: []metadataserver.Option{
				metadataserver.WithAddress("4.3.2.1"),
				metadataserver.WithPort(7777),
				metadataserver.WithHandlers(
					map[string]metadataserver.Metadata{
						"entry1": func() string {
							return "one"
						},
					}),
			},
			want: metadataserver.Configuration{
				Address:         "4.3.2.1",
				Endpoint:        metadataserver.DefaultEndpoint,
				Port:            7777,
				ShutdownTimeout: metadataserver.DefaultShutdownTimeout,
				Handlers: map[string]metadataserver.Metadata{
					"entry1": func() string {
						return "one"
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := metadataserver.New(test.input...)
			if err != nil {
				t.Errorf("expected no errors, got: %v", err)
			}
			if diff := cmp.Diff(test.want, got.Configuration(), opt); diff != "" {
				t.Errorf("handlers mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestHandlers(t *testing.T) {
	testConfig := &metadataserver.Configuration{
		Address:         "1.2.3.4",
		Endpoint:        "/custom/endpoint",
		Port:            8080,
		ShutdownTimeout: 15,
		Handlers: map[string]metadataserver.Metadata{
			"entry1": func() string {
				return "one"
			},
			"entry2": func() string {
				return "two"
			},
		},
	}
	s, err := metadataserver.New(metadataserver.WithConfiguration(testConfig))
	if err != nil {
		t.Errorf("expected no errors, got: %v", err)
	}
	ts := httptest.NewServer(s.HttpHandler())
	defer ts.Close()
	for e, want := range testConfig.Handlers {
		ep := path.Join(s.Configuration().Endpoint, e)
		res, err := http.Get(ts.URL + ep)
		if err != nil {
			t.Errorf("expected no errors, got: %v", err)
		}
		got, err := io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			t.Errorf("expected no errors, got: %v", err)
		}
		if diff := cmp.Diff(want(), string(got)); diff != "" {
			t.Errorf("server response mismatch (-want +got):\n%s", diff)
		}
	}
}

func TestEndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	s, err := startLiveServer(context.Background(), "0.0.0.0")
	if err != nil {
		t.Errorf("expected no errors, got: %v", err)
	}
	url := fmt.Sprintf("http://127.0.0.1:%d/computeMetadata/v1", s.Configuration().Port)
	res, err := http.Get(url)
	if err != nil {
		t.Errorf("expected no errors, got: %v", err)
	}
	defer res.Body.Close()
	resData, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("expected no errors, got: %v", err)
	}
	if string(resData) != "ok" {
		t.Errorf("expected:\"ok\", got: %v", err)
	}
	if err := s.Stop(context.Background()); err != nil {
		t.Errorf("expected no errors, got: %v", err)
	}
}

func TestAlreadyStartedServer(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	s, err := startLiveServer(context.Background(), "0.0.0.0")
	if err != nil {
		t.Errorf("expected no errors, got: %v", err)
	}
	if err := s.Start(context.Background()); err != metadataserver.ErrServerAlreadyStarted {
		t.Errorf("expected error: %v", metadataserver.ErrServerAlreadyStarted)
	}
	if err := s.Stop(context.Background()); err != nil {
		t.Errorf("expected no errors, got: %v", err)
	}
}

func TestStopNotRunningServer(t *testing.T) {
	s, err := metadataserver.New(
		metadataserver.WithAddress("0.0.0.0"),
		metadataserver.WithPort(8182))
	if err != nil {
		t.Errorf("expected no errors, got: %v", err)
	}
	if err := s.Stop(context.Background()); err != metadataserver.ErrServerIsNotRunning {
		t.Errorf("expected error: %v", metadataserver.ErrServerIsNotRunning)
	}
}

func startLiveServer(ctx context.Context, ip string) (*metadataserver.Server, error) {
	var port int
	dummy := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	server := httptest.NewServer(dummy)
	fmt.Sscanf(server.URL, "http://127.0.0.1:%d", &port)
	server.Close()

	s, err := metadataserver.New(
		metadataserver.WithAddress(ip),
		metadataserver.WithPort(port))
	if err != nil {
		return s, err
	}
	err = s.Start(ctx)
	return s, err
}
