package config

import (
	"os"
	"testing"
)

func TestDefaults(t *testing.T) {
	cfg := defaults()

	if cfg.Server.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.Server.Port)
	}

	if cfg.Logging.Level != "info" {
		t.Errorf("expected default log level 'info', got %s", cfg.Logging.Level)
	}

	if cfg.Service.Name != "gogovcode" {
		t.Errorf("expected service name 'gogovcode', got %s", cfg.Service.Name)
	}
}

func TestLoadFromEnv(t *testing.T) {
	// Save original env
	originalPort := os.Getenv("GOGOVCODE_PORT")
	originalLevel := os.Getenv("GOGOVCODE_LOG_LEVEL")

	// Set test env vars
	os.Setenv("GOGOVCODE_PORT", "9000")
	os.Setenv("GOGOVCODE_LOG_LEVEL", "debug")

	// Restore original env after test
	defer func() {
		os.Setenv("GOGOVCODE_PORT", originalPort)
		os.Setenv("GOGOVCODE_LOG_LEVEL", originalLevel)
	}()

	cfg := defaults()
	loadFromEnv(cfg)

	if cfg.Server.Port != 9000 {
		t.Errorf("expected port 9000 from env, got %d", cfg.Server.Port)
	}

	if cfg.Logging.Level != "debug" {
		t.Errorf("expected log level 'debug' from env, got %s", cfg.Logging.Level)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name:    "valid config",
			cfg:     defaults(),
			wantErr: false,
		},
		{
			name: "invalid port - too low",
			cfg: &Config{
				Server:  ServerConfig{Port: 0},
				Logging: LoggingConfig{Level: "info", Format: "json"},
			},
			wantErr: true,
		},
		{
			name: "invalid port - too high",
			cfg: &Config{
				Server:  ServerConfig{Port: 99999},
				Logging: LoggingConfig{Level: "info", Format: "json"},
			},
			wantErr: true,
		},
		{
			name: "invalid log level",
			cfg: &Config{
				Server:  ServerConfig{Port: 8080},
				Logging: LoggingConfig{Level: "invalid", Format: "json"},
			},
			wantErr: true,
		},
		{
			name: "tls enabled without cert",
			cfg: &Config{
				Server:  ServerConfig{Port: 8080},
				TLS:     TLSConfig{Enabled: true},
				Logging: LoggingConfig{Level: "info", Format: "json"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAddr(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
	}

	expected := "localhost:8080"
	if addr := cfg.Addr(); addr != expected {
		t.Errorf("expected addr %s, got %s", expected, addr)
	}
}

func TestApplyProfileDefaults(t *testing.T) {
	tests := []struct {
		name          string
		profile       Profile
		expectedLevel string
		expectedTLS   bool
	}{
		{
			name:          "dev profile",
			profile:       ProfileDev,
			expectedLevel: "debug",
			expectedTLS:   false,
		},
		{
			name:          "test profile",
			profile:       ProfileTest,
			expectedLevel: "info",
			expectedTLS:   false,
		},
		{
			name:          "prod profile",
			profile:       ProfileProd,
			expectedLevel: "warn",
			expectedTLS:   false,
		},
		{
			name:          "dsmil profile",
			profile:       ProfileDSMIL,
			expectedLevel: "info",
			expectedTLS:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Profile: tt.profile,
				Logging: LoggingConfig{},
				TLS:     TLSConfig{},
			}

			applyProfileDefaults(cfg)

			if cfg.Logging.Level != tt.expectedLevel {
				t.Errorf("expected log level %s, got %s", tt.expectedLevel, cfg.Logging.Level)
			}

			if cfg.TLS.Enabled != tt.expectedTLS {
				t.Errorf("expected TLS enabled %v, got %v", tt.expectedTLS, cfg.TLS.Enabled)
			}
		})
	}
}
