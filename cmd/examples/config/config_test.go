package config_test

import (
	"testing"

	"github.com/opensrehub/otelgokit/cmd/examples/config"
)

func TestLoadConfigDefaults(t *testing.T) {
	cfg := config.LoadConfig()

	if cfg.ServiceName == "" {
		t.Error("ServiceName should not be empty")
	}
	if cfg.CollectorURL == "" {
		t.Error("CollectorURL should not be empty")
	}
	if cfg.Port == 0 {
		t.Error("Port should not be zero")
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	t.Setenv("SERVICE_NAME", "my-service")
	t.Setenv("ENVIRONMENT", "staging")
	t.Setenv("OTEL_COLLECTOR_URL", "collector:4317")
	t.Setenv("PORT", "8080")

	cfg := config.LoadConfig()

	if cfg.ServiceName != "my-service" {
		t.Errorf("ServiceName = %q, want %q", cfg.ServiceName, "my-service")
	}
	if cfg.Environment != "staging" {
		t.Errorf("Environment = %q, want %q", cfg.Environment, "staging")
	}
	if cfg.CollectorURL != "collector:4317" {
		t.Errorf("CollectorURL = %q, want %q", cfg.CollectorURL, "collector:4317")
	}
	if cfg.Port != 8080 {
		t.Errorf("Port = %d, want %d", cfg.Port, 8080)
	}
}

func TestLoadConfigInvalidPort(t *testing.T) {
	t.Setenv("PORT", "not-a-number")

	cfg := config.LoadConfig()

	if cfg.Port != 3000 {
		t.Errorf("Port = %d, want default 3000 on invalid input", cfg.Port)
	}
}
