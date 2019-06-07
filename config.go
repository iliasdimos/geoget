package main

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config is the base struct for our server's configuration
type Config struct {
	Log struct {
		Level string `envconfig:"LOG_LEVEL" default:"info"`
		Dev   bool   `envconfig:"DEV" default:"false"`
		Debug bool   `envconfig:"DEBUG" default:"false"`
	}
	Web struct {
		Host            string        `envconfig:"HOST" default:"0.0.0.0"`
		Port            string        `envconfig:"PORT" default:"8000"`
		ReadTimeout     time.Duration `envconfig:"READ_TIMEOUT" default:"10s"`
		WriteTimeout    time.Duration `envconfig:"WRITE_TIMEOUT" default:"20s"`
		ShutdownTimeout time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"10s"`
		Debug           bool          `envconfig:"DEBUG" default:"false"`
	}
	Database struct {
		Path string `envconfig:"DATABASE" default:"data/maxmind.mmdb"`
	}
}

// newCfg return a new config struct
func newCfg(prefix string) (*Config, error) {
	cfg := new(Config)
	err := envconfig.Process(prefix, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
