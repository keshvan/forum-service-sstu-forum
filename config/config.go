package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Env         string        `yaml:"env" env-default:"local"`
	PG_URL      string        `yaml:"pg_url"`
	AccessTTL   time.Duration `yaml:"access_ttl"`
	RefreshTTL  time.Duration `yaml:"refresh_ttl" env-required:"true"`
	Log         string        `yaml:"log_level"`
	Server      string        `yaml:"server"`
	Secret      string        `yaml:"secret"`
	GrpcAddress string        `yaml:"grpc_address"`
}

func NewConfig() (*Config, error) {
	cfg := &Config{}
	file, err := os.ReadFile("./config.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	if err := yaml.Unmarshal(file, cfg); err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	return cfg, nil
}
