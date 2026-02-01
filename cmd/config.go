package cmd

import (
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type RestConfig struct {
	Port     string `mapstructure:"port"`
	Timeouts struct {
		Read       time.Duration `mapstructure:"read"`
		ReadHeader time.Duration `mapstructure:"read_header"`
		Write      time.Duration `mapstructure:"write"`
		Idle       time.Duration `mapstructure:"idle"`
	} `mapstructure:"timeouts"`
}

type DatabaseConfig struct {
	SQL struct {
		URL string `mapstructure:"url"`
	} `mapstructure:"sql"`
}

type OAuthConfig struct {
	Google struct {
		ClientID     string `mapstructure:"client_id"`
		ClientSecret string `mapstructure:"client_secret"`
		RedirectURL  string `mapstructure:"redirect_url"`
	}
}

type KafkaConfig struct {
	Brokers []string `mapstructure:"brokers"`
}

type JWTConfig struct {
	User struct {
		AccessToken struct {
			SecretKey     string        `mapstructure:"secret_key"`
			TokenLifetime time.Duration `mapstructure:"token_lifetime"`
		} `mapstructure:"access_token"`
		RefreshToken struct {
			SecretKey     string        `mapstructure:"secret_key"`
			HashKey       string        `mapstructure:"hash_key"`
			TokenLifetime time.Duration `mapstructure:"token_lifetime"`
		} `mapstructure:"refresh_token"`
	} `mapstructure:"user"`
}

type Config struct {
	Log      LogConfig      `mapstructure:"log"`
	Rest     RestConfig     `mapstructure:"rest"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	OAuth    OAuthConfig    `mapstructure:"oauth"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
	Database DatabaseConfig `mapstructure:"database"`
}

func LoadConfig() (Config, error) {
	configPath := os.Getenv("KV_VIPER_FILE")
	if configPath == "" {
		return Config{}, errors.New("KV_VIPER_FILE env var is not set")
	}
	viper.SetConfigFile(configPath)

	if err := viper.ReadInConfig(); err != nil {
		return Config{}, errors.Errorf("error reading config file: %s", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return Config{}, errors.Errorf("error unmarshalling config: %s", err)
	}

	return config, nil
}

func (c *Config) GoogleOAuth() oauth2.Config {
	return oauth2.Config{
		ClientID:     c.OAuth.Google.ClientID,
		ClientSecret: c.OAuth.Google.ClientSecret,
		RedirectURL:  c.OAuth.Google.RedirectURL,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
}
