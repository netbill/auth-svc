package config

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/netbill/auth-svc/pkg/log"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type ServiceCfg struct {
	Name string `mapstructure:"name"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type DatabaseConfig struct {
	SQL struct {
		URL string `mapstructure:"url"`
	} `mapstructure:"sql"`
}

type RestConfig struct {
	Port     int `mapstructure:"port"`
	Timeouts struct {
		Read       time.Duration `mapstructure:"read"`
		ReadHeader time.Duration `mapstructure:"read_header"`
		Write      time.Duration `mapstructure:"write"`
		Idle       time.Duration `mapstructure:"idle"`
	} `mapstructure:"timeouts"`
}

type AuthConfig struct {
	Tokens struct {
		Issuer        string `mapstructure:"issuer"`
		AccountAccess struct {
			SecretKey string        `mapstructure:"secret_key"`
			TTL       time.Duration `mapstructure:"ttl"`
		} `mapstructure:"account_access"`
		AccountRefresh struct {
			SecretKey string        `mapstructure:"secret_key"`
			HashKey   string        `mapstructure:"hash_key"`
			TTL       time.Duration `mapstructure:"ttl"`
		} `mapstructure:"account_refresh"`
	} `mapstructure:"tokens"`

	OAuth struct {
		Google struct {
			ClientID     string `mapstructure:"client_id"`
			ClientSecret string `mapstructure:"client_secret"`
			RedirectURL  string `mapstructure:"redirect_url"`
		} `mapstructure:"google"`
	} `mapstructure:"oauth"`
}

type KafkaConfig struct {
	Brokers  []string `mapstructure:"brokers"`
	Identity string   `mapstructure:"identity"`

	Produce struct {
		Topics struct {
			AccountsV1 struct {
				RequiredAcks string        `mapstructure:"required_acks"`
				Compression  string        `mapstructure:"compression"`
				Balancer     string        `mapstructure:"balancer"`
				BatchSize    int           `mapstructure:"batch_size"`
				BatchTimeout time.Duration `mapstructure:"batch_timeout"`
				DialTimeout  time.Duration `mapstructure:"dial_timeout"`
				IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
			} `mapstructure:"accounts_v1"`
		} `mapstructure:"topics"`
	} `mapstructure:"produce"`

	Consume struct {
		Backoff struct {
			Min time.Duration `mapstructure:"min"`
			Max time.Duration `mapstructure:"max"`
		} `mapstructure:"backoff"`

		Topics struct {
			OrganizationsV1 struct {
				Instances      int           `mapstructure:"instances"`
				MinBytes       int           `mapstructure:"min_bytes"`
				MaxBytes       int           `mapstructure:"max_bytes"`
				MaxWait        time.Duration `mapstructure:"max_wait"`
				CommitInterval time.Duration `mapstructure:"commit_interval"`
				QueueCapacity  int           `mapstructure:"queue_capacity"`
			} `mapstructure:"organizations_v1"`

			OrgMembersV1 struct {
				Instances      int           `mapstructure:"instances"`
				MinBytes       int           `mapstructure:"min_bytes"`
				MaxBytes       int           `mapstructure:"max_bytes"`
				MaxWait        time.Duration `mapstructure:"max_wait"`
				CommitInterval time.Duration `mapstructure:"commit_interval"`
				QueueCapacity  int           `mapstructure:"queue_capacity"`
			} `mapstructure:"organization_members_v1"`
		} `mapstructure:"topics"`
	} `mapstructure:"consume"`

	Inbox struct {
		Routines       int           `mapstructure:"routines"`
		Slots          int           `mapstructure:"slots"`
		BatchSize      int           `mapstructure:"batch_size"`
		Sleep          time.Duration `mapstructure:"sleep"`
		MinNextAttempt time.Duration `mapstructure:"min_next_attempt"`
		MaxNextAttempt time.Duration `mapstructure:"max_next_attempt"`
		MaxAttempts    int32         `mapstructure:"max_attempts"`
	} `mapstructure:"inbox"`

	Outbox struct {
		Routines       int           `mapstructure:"routines"`
		Slots          int           `mapstructure:"slots"`
		BatchSize      int           `mapstructure:"batch_size"`
		Sleep          time.Duration `mapstructure:"sleep"`
		MinNextAttempt time.Duration `mapstructure:"min_next_attempt"`
		MaxNextAttempt time.Duration `mapstructure:"max_next_attempt"`
		MaxAttempts    int32         `mapstructure:"max_attempts"`
	} `mapstructure:"outbox"`
}

type Config struct {
	Service  ServiceCfg     `mapstructure:"service"`
	Log      LogConfig      `mapstructure:"log"`
	Database DatabaseConfig `mapstructure:"database"`
	Rest     RestConfig     `mapstructure:"rest"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
}

func LoadConfig() *Config {
	configPath := os.Getenv("KV_VIPER_FILE")
	if configPath == "" {
		panic(fmt.Errorf("KV_VIPER_FILE env var is not set"))
	}
	viper.SetConfigFile(configPath)

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("error reading config file: %s", err))
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		panic(fmt.Errorf("error unmarshalling config: %s", err))
	}

	return &config
}

func (cfg *Config) Logger() *log.Logger {
	return log.New(cfg.Log.Level, cfg.Log.Format, cfg.Service.Name)
}

func (cfg *Config) PoolDB(ctx context.Context) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, cfg.Database.SQL.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return pool, nil
}

func (cfg *Config) GoogleOAuth() oauth2.Config {
	return oauth2.Config{
		ClientID:     cfg.Auth.OAuth.Google.ClientID,
		ClientSecret: cfg.Auth.OAuth.Google.ClientSecret,
		RedirectURL:  cfg.Auth.OAuth.Google.RedirectURL,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
}
