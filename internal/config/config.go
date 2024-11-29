package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

var _prefix string

// SetPrefix sets the prefix used for all environment variable names used by this service.
func SetPrefix(prefix string) {
	_prefix = prefix
}

// Config holds the application configuration settings.
type Config struct {
	Service  Service
	Database Database
	HTTP     HTTP
	GRPC     GRPC
}

// Service holds configuration settings for the application service.
type Service struct {
	Name     string `envconfig:"SERVICE_NAME" required:"true"`
	Version  string `envconfig:"SERVICE_VERSION" default:"0.0.0-dev"`
	LogLevel string `envconfig:"SERVICE_LOG_LEVEL" default:"info"`
}

// Database holds the configuration for connecting to a database.
type Database struct {
	Host string `envconfig:"DB_HOST" default:"localhost"`
	Port string `envconfig:"DB_PORT" default:"5432"`
	User string `envconfig:"DB_USER" required:"true"`
	Pass string `envconfig:"DB_PASS" required:"true"`
	Name string `envconfig:"DB_NAME" required:"true"`
}

// HTTP represents the configuration required for an HTTP server.
type HTTP struct {
	Port string `envconfig:"HTTP_PORT" default:"8080"`
}

// GRPC represents a gRPC server configuration.
type GRPC struct {
	Port string `envconfig:"GRPC_PORT" default:"9090"`
}

// Load loads environment variables from the specified files and processes them into a Config struct.
func Load(envFiles ...string) (*Config, error) {
	// this will load from local env files if they exist
	_ = godotenv.Overload(envFiles...)

	var out Config
	if err := envconfig.Process(_prefix, &out); err != nil {
		return nil, err
	}

	return &out, nil
}
