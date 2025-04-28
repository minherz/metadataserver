package metadataserver_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/minherz/metadataserver"
)

var opt = cmp.Comparer(func(x, y metadataserver.Metadata) bool {
	return x() == y()
})

func TestNewConfiguration(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]metadataserver.Metadata
		want  map[string]metadataserver.Metadata
	}{
		{
			name:  "nil",
			input: nil,
			want:  metadataserver.DefaultConfigurationHandlers,
		},
		{
			name:  "empty_handers",
			input: map[string]metadataserver.Metadata{},
			want:  metadataserver.DefaultConfigurationHandlers,
		},
		{
			name:  "default_handlers",
			input: metadataserver.DefaultConfigurationHandlers,
			want:  metadataserver.DefaultConfigurationHandlers,
		},
		{
			name: "custom_handlers",
			input: map[string]metadataserver.Metadata{
				"custom1": func() string { return "value1" },
				"custom2": func() string { return "value2" },
			},
			want: map[string]metadataserver.Metadata{
				"custom1": func() string { return "value1" },
				"custom2": func() string { return "value2" },
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := metadataserver.NewConfiguration(test.input)
			if diff := cmp.Diff(test.want, got.Handlers, opt); diff != "" {
				t.Errorf("configuration mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestNewConfigFromFile(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  *metadataserver.Configuration
	}{
		{
			name:  "smoke_test",
			input: "test/fixtures/config_smoke_test.json",
			want: &metadataserver.Configuration{
				Address:         "1.2.3.4",
				Endpoint:        "/custom/endpoint",
				Port:            8080,
				ShutdownTimeout: 15,
				Handlers: map[string]metadataserver.Metadata{
					"entry1": func() string {
						return "one"
					},
				},
			},
		},
		{
			name:  "handlers_only",
			input: "test/fixtures/config_handlers.json",
			want: &metadataserver.Configuration{
				Address:         metadataserver.DefaultAddress,
				Endpoint:        metadataserver.DefaultEndpoint,
				Port:            metadataserver.DefaultPort,
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
			got, err := metadataserver.NewConfigFromFile(test.input)
			if err != nil {
				t.Errorf("expected no errors:\n%s", err)
			}
			if diff := cmp.Diff(test.want, got, opt); diff != "" {
				t.Errorf("configutation mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func setupTestConfigureEnvHandlersFromFile(tb testing.TB) {
	tb.Setenv("two", "handler_from_env_var")
}

func TestConfigureEnvHandlersFromFile(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  map[string]metadataserver.Metadata
	}{
		{
			name:  "env_handlers",
			input: "test/fixtures/config_env_handlers.json",
			want: map[string]metadataserver.Metadata{
				"entry2": func() string { return "handler_from_env_var" },
			},
		},
		{
			name:  "mixed_handlers",
			input: "test/fixtures/config_mixed_handlers.json",
			want: map[string]metadataserver.Metadata{
				"entry1": func() string { return "one" },
				"entry2": func() string { return "handler_from_env_var" },
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setupTestConfigureEnvHandlersFromFile(t)

			got, err := metadataserver.NewConfigFromFile(test.input)
			if err != nil {
				t.Errorf("expected no errors:\n%s", err)
			}
			if diff := cmp.Diff(test.want, got.Handlers, opt); diff != "" {
				t.Errorf("handlers mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
