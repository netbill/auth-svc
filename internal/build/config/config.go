package config

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/netbill/auth-svc/pkg/log"
	"github.com/redis/go-redis/v9"
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

	Redis struct {
		Addr     string `mapstructure:"addr"`
		Password string `mapstructure:"password"`
		DB       int    `mapstructure:"db"`

		TTL struct {
			Account  time.Duration `mapstructure:"account"`
			Email    time.Duration `mapstructure:"email"`
			Password time.Duration `mapstructure:"password"`
			Session  time.Duration `mapstructure:"session"`
		} `mapstructure:"ttl"`
	} `mapstructure:"redis"`
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

	PassBcryptCost int `mapstructure:"pass_bcrypt_cost"`
}

type KafkaConfig struct {
	Brokers  []string `mapstructure:"brokers"`
	Identity string   `mapstructure:"identity"`
}

type GRPCConfig struct {
	Port int `mapstructure:"port"`
}

type OTELConfig struct {
	CollectorEndpoint string  `mapstructure:"collector_endpoint"`
	SamplingRatio     float64 `mapstructure:"sampling_ratio"`
}

type Config struct {
	Service  ServiceCfg     `mapstructure:"service"`
	Log      LogConfig      `mapstructure:"log"`
	Database DatabaseConfig `mapstructure:"database"`
	Rest     RestConfig     `mapstructure:"rest"`
	GRPC     GRPCConfig     `mapstructure:"grpc"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
	OTEL     OTELConfig     `mapstructure:"otel"`
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

	for _, key := range viper.AllKeys() {
		if val, ok := viper.Get(key).(string); ok {
			viper.Set(key, os.ExpandEnv(val))
		}
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
	// ParseConfig отдельно от New — нужно чтобы добавить Tracer до создания коннектов.
	// pgxpool.New внутри тоже вызывает ParseConfig, но tracer туда не пробросить.
	poolCfg, err := pgxpool.ParseConfig(cfg.Database.SQL.URL)
	if err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}

	// otelpgx.NewTracer() берёт глобальный TracerProvider (установлен в telemetry.Init).
	// Каждый SQL запрос получает child span с именем "SELECT", "INSERT" и т.д.,
	// временем выполнения и атрибутами db.statement, db.operation.
	poolCfg.ConnConfig.Tracer = otelpgx.NewTracer()

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return pool, nil
}

func (cfg *Config) RedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.Database.Redis.Addr,
		Password: cfg.Database.Redis.Password,
		DB:       cfg.Database.Redis.DB,
	})
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
