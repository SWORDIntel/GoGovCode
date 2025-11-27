package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

// Profile represents the deployment environment
type Profile string

const (
	ProfileDev   Profile = "dev"
	ProfileTest  Profile = "test"
	ProfileProd  Profile = "prod"
	ProfileDSMIL Profile = "dsmil"
)

// Config holds all configuration for GoGovCode
type Config struct {
	// Server configuration
	Server ServerConfig `json:"server"`

	// TLS configuration
	TLS TLSConfig `json:"tls"`

	// Logging configuration
	Logging LoggingConfig `json:"logging"`

	// Redis configuration (placeholder for future phases)
	Redis RedisConfig `json:"redis"`

	// MinIO configuration (placeholder for future phases)
	MinIO MinIOConfig `json:"minio"`

	// Service metadata
	Service ServiceConfig `json:"service"`

	// Profile
	Profile Profile `json:"profile"`
}

// ServerConfig holds HTTP server settings
type ServerConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// TLSConfig holds TLS/HTTPS settings
type TLSConfig struct {
	Enabled  bool   `json:"enabled"`
	CertFile string `json:"cert_file"`
	KeyFile  string `json:"key_file"`
}

// LoggingConfig holds logging settings
type LoggingConfig struct {
	Level  string `json:"level"`  // debug, info, warn, error
	Format string `json:"format"` // json, text
}

// RedisConfig holds Redis connection settings
type RedisConfig struct {
	Enabled  bool   `json:"enabled"`
	Endpoint string `json:"endpoint"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

// MinIOConfig holds MinIO connection settings
type MinIOConfig struct {
	Enabled   bool   `json:"enabled"`
	Endpoint  string `json:"endpoint"`
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	Bucket    string `json:"bucket"`
	UseSSL    bool   `json:"use_ssl"`
}

// ServiceConfig holds service metadata
type ServiceConfig struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Load loads configuration from file, environment, and flags
// Priority: flags > env > file > defaults
func Load() (*Config, error) {
	cfg := defaults()

	// Parse command-line flags
	configFile := flag.String("config", "", "Path to configuration file")
	profile := flag.String("profile", string(ProfileDev), "Deployment profile (dev|test|prod|dsmil)")
	host := flag.String("host", "", "Server host")
	port := flag.Int("port", 0, "Server port")
	logLevel := flag.String("log-level", "", "Log level (debug|info|warn|error)")
	tlsEnabled := flag.Bool("tls", false, "Enable TLS")

	flag.Parse()

	// Set profile
	cfg.Profile = Profile(*profile)

	// Load from config file if provided
	if *configFile != "" {
		if err := loadFromFile(*configFile, cfg); err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}

	// Override with environment variables
	loadFromEnv(cfg)

	// Override with command-line flags
	if *host != "" {
		cfg.Server.Host = *host
	}
	if *port != 0 {
		cfg.Server.Port = *port
	}
	if *logLevel != "" {
		cfg.Logging.Level = *logLevel
	}
	if *tlsEnabled {
		cfg.TLS.Enabled = true
	}

	// Apply profile-specific defaults
	applyProfileDefaults(cfg)

	return cfg, nil
}

// defaults returns default configuration
func defaults() *Config {
	return &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
		TLS: TLSConfig{
			Enabled:  false,
			CertFile: "",
			KeyFile:  "",
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
		Redis: RedisConfig{
			Enabled:  false,
			Endpoint: "localhost:6379",
			Password: "",
			DB:       0,
		},
		MinIO: MinIOConfig{
			Enabled:   false,
			Endpoint:  "localhost:9000",
			AccessKey: "",
			SecretKey: "",
			Bucket:    "audit",
			UseSSL:    false,
		},
		Service: ServiceConfig{
			Name:    "gogovcode",
			Version: "1.0.0-phase1",
		},
		Profile: ProfileDev,
	}
}

// loadFromFile loads configuration from a JSON file
func loadFromFile(path string, cfg *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, cfg)
}

// loadFromEnv loads configuration from environment variables
func loadFromEnv(cfg *Config) {
	if v := os.Getenv("GOGOVCODE_HOST"); v != "" {
		cfg.Server.Host = v
	}
	if v := os.Getenv("GOGOVCODE_PORT"); v != "" {
		var port int
		fmt.Sscanf(v, "%d", &port)
		if port > 0 {
			cfg.Server.Port = port
		}
	}
	if v := os.Getenv("GOGOVCODE_LOG_LEVEL"); v != "" {
		cfg.Logging.Level = strings.ToLower(v)
	}
	if v := os.Getenv("GOGOVCODE_LOG_FORMAT"); v != "" {
		cfg.Logging.Format = strings.ToLower(v)
	}
	if v := os.Getenv("GOGOVCODE_TLS_ENABLED"); v == "true" || v == "1" {
		cfg.TLS.Enabled = true
	}
	if v := os.Getenv("GOGOVCODE_TLS_CERT"); v != "" {
		cfg.TLS.CertFile = v
	}
	if v := os.Getenv("GOGOVCODE_TLS_KEY"); v != "" {
		cfg.TLS.KeyFile = v
	}
	if v := os.Getenv("GOGOVCODE_REDIS_ENABLED"); v == "true" || v == "1" {
		cfg.Redis.Enabled = true
	}
	if v := os.Getenv("GOGOVCODE_REDIS_ENDPOINT"); v != "" {
		cfg.Redis.Endpoint = v
	}
	if v := os.Getenv("GOGOVCODE_REDIS_PASSWORD"); v != "" {
		cfg.Redis.Password = v
	}
	if v := os.Getenv("GOGOVCODE_MINIO_ENABLED"); v == "true" || v == "1" {
		cfg.MinIO.Enabled = true
	}
	if v := os.Getenv("GOGOVCODE_MINIO_ENDPOINT"); v != "" {
		cfg.MinIO.Endpoint = v
	}
	if v := os.Getenv("GOGOVCODE_MINIO_ACCESS_KEY"); v != "" {
		cfg.MinIO.AccessKey = v
	}
	if v := os.Getenv("GOGOVCODE_MINIO_SECRET_KEY"); v != "" {
		cfg.MinIO.SecretKey = v
	}
	if v := os.Getenv("GOGOVCODE_SERVICE_NAME"); v != "" {
		cfg.Service.Name = v
	}
	if v := os.Getenv("GOGOVCODE_SERVICE_VERSION"); v != "" {
		cfg.Service.Version = v
	}
}

// applyProfileDefaults applies profile-specific defaults
func applyProfileDefaults(cfg *Config) {
	switch cfg.Profile {
	case ProfileDev:
		// Development: verbose logging, no TLS
		if cfg.Logging.Level == "" {
			cfg.Logging.Level = "debug"
		}
		cfg.TLS.Enabled = false

	case ProfileTest:
		// Test: info logging, no TLS
		if cfg.Logging.Level == "" {
			cfg.Logging.Level = "info"
		}
		cfg.TLS.Enabled = false

	case ProfileProd:
		// Production: warn logging, TLS recommended
		if cfg.Logging.Level == "" {
			cfg.Logging.Level = "warn"
		}

	case ProfileDSMIL:
		// DSMIL: info logging, TLS required, all security features enabled
		if cfg.Logging.Level == "" {
			cfg.Logging.Level = "info"
		}
		cfg.TLS.Enabled = true
		// Future phases will enable additional security features here
	}
}

// Addr returns the server address as host:port
func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.TLS.Enabled {
		if c.TLS.CertFile == "" || c.TLS.KeyFile == "" {
			return fmt.Errorf("TLS enabled but cert/key files not specified")
		}
	}

	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[c.Logging.Level] {
		return fmt.Errorf("invalid log level: %s", c.Logging.Level)
	}

	validFormats := map[string]bool{"json": true, "text": true}
	if !validFormats[c.Logging.Format] {
		return fmt.Errorf("invalid log format: %s", c.Logging.Format)
	}

	return nil
}
