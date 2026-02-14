package boot

import (
	"fmt"
	"os"

	"github.com/netbill/auth-svc/internal/messenger"
	"github.com/netbill/auth-svc/internal/rest"
	"github.com/netbill/auth-svc/internal/tokenmanger"
	"github.com/spf13/viper"
)

const ServiceName = "auth-svc"

type DatabaseConfig struct {
	SQL struct {
		URL string `mapstructure:"url"`
	} `mapstructure:"sql"`
}

type AuthConfig struct {
	Tokens tokenmanger.Config `mapstructure:"tokens"`
	OAuth  OAuthConfig        `mapstructure:"oauth"`
}

type Config struct {
	Log      LogConfig        `mapstructure:"log"`
	Rest     rest.Config      `mapstructure:"rest"`
	Auth     AuthConfig       `mapstructure:"auth"`
	Kafka    messenger.Config `mapstructure:"kafka"`
	Database DatabaseConfig   `mapstructure:"database"`
}

func LoadConfig() (*Config, error) {
	configPath := os.Getenv("KV_VIPER_FILE")
	if configPath == "" {
		return nil, fmt.Errorf("KV_VIPER_FILE env var is not set")
	}

	viper.SetConfigFile(configPath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %s", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %s", err)
	}

	return &config, nil
}
