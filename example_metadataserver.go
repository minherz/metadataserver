package metadataserver

import (
	"context"
	"log/slog"
	"os"
)

func ExampleWithConfiguration() {
	c := NewConfiguration(map[string]Metadata{
		"project-id": func() string {
			return os.Getenv("GOOGLE_CLOUD_PROJECT")
		},
		"instance/zone": func() string {
			return "us-central1"
		},
	})
	s, err := New(WithConfiguration(c))
	if err != nil {
		// handle error
	}
	s.Start(context.Background())
}

func ExampleWithLogger() {
	s, err := New(WithLogger(slog.Default()))
	if err != nil {
		// handle error
	}
	s.Start(context.Background())
}

func ExampleNew() {
	s, err := New(WithAddress("0.0.0.0"), WithPort(8080))
	if err != nil {
		// handle error
	}
	s.Start(context.Background())
}
