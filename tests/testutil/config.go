package testutil

import (
	"os"
	"time"

	"go.yaml.in/yaml/v3"
)

type Config struct {
	Database struct {
		SQL struct {
			URL string `yaml:"url"`
		} `yaml:"sql"`
		Redis struct {
			Addr     string `yaml:"addr"`
			Password string `yaml:"password"`
			DB       int    `yaml:"db"`
		} `yaml:"redis"`
	} `yaml:"database"`

	Auth struct {
		Tokens struct {
			Issuer     string `yaml:"issuer"`
			UserAccess struct {
				SecretKey string        `yaml:"secret_key"`
				TTL       time.Duration `yaml:"ttl"`
			} `yaml:"user_access"`
			UserRefresh struct {
				SecretKey string        `yaml:"secret_key"`
				HashKey   string        `yaml:"hash_key"`
				TTL       time.Duration `yaml:"ttl"`
			} `yaml:"user_refresh"`
		} `yaml:"tokens"`
		PassBcryptCost int `yaml:"pass_bcrypt_cost"`
	} `yaml:"auth"`

	GRPC struct {
		Addr string `yaml:"addr"`
	} `yaml:"grpc"`

	REST struct {
		BaseURL string `yaml:"base_url"`
	} `yaml:"rest"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
