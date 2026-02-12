package config

import (
	"fmt"
	"os"
	"time"

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

type KafkaConfig struct {
	Brokers []string `mapstructure:"brokers"`
	Readers struct {
		OrgMemberV1 int `mapstructure:"organizations_member_v1"`
	} `mapstructure:"readers"`
	Inbox struct {
		ProcessCount   int           `mapstructure:"process_count"`
		Routines       int           `mapstructure:"routines"`
		MinBatch       int           `mapstructure:"min_batch"`
		MaxBatch       int           `mapstructure:"max_batch"`
		MinSleep       time.Duration `mapstructure:"min_sleep"`
		MaxSleep       time.Duration `mapstructure:"max_sleep"`
		MinNextAttempt time.Duration `mapstructure:"min_next_attempt"`
		MaxNextAttempt time.Duration `mapstructure:"max_next_attempt"`
		MaxAttempts    int32         `mapstructure:"max_attempts"`
	} `mapstructure:"inbox"`
	Outbox struct {
		ProcessCount   int           `mapstructure:"process_count"`
		Routines       int           `mapstructure:"routines"`
		MinBatch       int           `mapstructure:"min_batch"`
		MaxBatch       int           `mapstructure:"max_batch"`
		MinSleep       time.Duration `mapstructure:"min_sleep"`
		MaxSleep       time.Duration `mapstructure:"max_sleep"`
		MinNextAttempt time.Duration `mapstructure:"min_next_attempt"`
		MaxNextAttempt time.Duration `mapstructure:"max_next_attempt"`
		MaxAttempts    int32         `mapstructure:"max_attempts"`
	} `mapstructure:"outbox"`
}

type AuthConfig struct {
	Account struct {
		Token struct {
			Access struct {
				SecretKey string        `mapstructure:"secret_key"`
				Lifetime  time.Duration `mapstructure:"lifetime"`
			} `mapstructure:"access"`
			Refresh struct {
				SecretKey string        `mapstructure:"secret_key"`
				HashKey   string        `mapstructure:"hash_key"`
				Lifetime  time.Duration `mapstructure:"lifetime"`
			} `mapstructure:"refresh"`
		} `mapstructure:"token"`
	} `mapstructure:"account"`

	OAuthConfig struct {
		Google struct {
			ClientID     string `mapstructure:"client_id"`
			ClientSecret string `mapstructure:"client_secret"`
			RedirectURL  string `mapstructure:"redirect_url"`
		} `mapstructure:"google"`
	} `mapstructure:"oauth"`
}

type Config struct {
	Log      LogConfig      `mapstructure:"log"`
	Rest     RestConfig     `mapstructure:"rest"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
	Database DatabaseConfig `mapstructure:"database"`
}

func LoadConfig() (Config, error) {
	configPath := os.Getenv("KV_VIPER_FILE")
	if configPath == "" {
		return Config{}, fmt.Errorf("KV_VIPER_FILE env var is not set")
	}

	viper.SetConfigFile(configPath)

	if err := viper.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("error reading config file: %s", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return Config{}, fmt.Errorf("error unmarshalling config: %s", err)
	}

	return config, nil
}

func (c *Config) GoogleOAuth() oauth2.Config {
	return oauth2.Config{
		ClientID:     c.Auth.OAuthConfig.Google.ClientID,
		ClientSecret: c.Auth.OAuthConfig.Google.ClientSecret,
		RedirectURL:  c.Auth.OAuthConfig.Google.RedirectURL,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
}
