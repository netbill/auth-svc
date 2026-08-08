package config

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/netbill/auth-svc/pkg/log"
	"github.com/netbill/awsx"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type ServiceCfg struct {
	Name string
}

type LogConfig struct {
	Level  string
	Format string
}

type SQLConfig struct {
	URL string
}

type RedisTTLConfig struct {
	User     time.Duration
	Email    time.Duration
	Password time.Duration
	Session  time.Duration
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	TTL      RedisTTLConfig
}

type DatabaseConfig struct {
	SQL   SQLConfig
	Redis RedisConfig
}

type RestTimeouts struct {
	Read       time.Duration
	ReadHeader time.Duration
	Write      time.Duration
	Idle       time.Duration
}

type RestConfig struct {
	Port     int
	Timeouts RestTimeouts
}

type TokenConfig struct {
	SecretKey string
	TTL       time.Duration
}

type RefreshTokenConfig struct {
	SecretKey string
	HashKey   string
	TTL       time.Duration
}

type AuthTokensConfig struct {
	Issuer      string
	UserAccess  TokenConfig
	UserRefresh RefreshTokenConfig
}

type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type AuthOAuthConfig struct {
	Google GoogleOAuthConfig
}

type AuthConfig struct {
	Tokens         AuthTokensConfig
	OAuth          AuthOAuthConfig
	PassBcryptCost int
}

type KafkaConfig struct {
	Brokers  []string
	Identity string
}

type GRPCConfig struct {
	Port int
}

type OTELConfig struct {
	CollectorEndpoint string
	SamplingRatio     float64
}

type S3AwsConfig struct {
	Region          string
	BucketName      string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	BaseURL         string
}

type S3MediaUserConfig struct {
	Avatar awsx.ImageValidator
}

type S3MediaResourcesConfig struct {
	User S3MediaUserConfig
}

type S3MediaLinkConfig struct {
	TTL time.Duration
}

type S3MediaConfig struct {
	Link      S3MediaLinkConfig
	Resources S3MediaResourcesConfig
}

type S3Config struct {
	Aws   S3AwsConfig
	Media S3MediaConfig
}

type Config struct {
	Service  ServiceCfg
	Log      LogConfig
	Database DatabaseConfig
	Rest     RestConfig
	GRPC     GRPCConfig
	Auth     AuthConfig
	S3       S3Config
	Kafka    KafkaConfig
	OTEL     OTELConfig
}

// LoadConfig builds the config entirely from environment variables — no
// config file involved. Vars backing an actual secret or a connection the
// service can't run without (DB, Redis, JWT keys) are required and panic if
// missing; everything else falls back to the same defaults the old
// config.example.yaml used to hardcode.
func LoadConfig() *Config {
	return &Config{
		Service: ServiceCfg{
			Name: envOr("SERVICE_NAME", "auth-svc"),
		},
		Log: LogConfig{
			Level:  envOr("LOG_LEVEL", "debug"),
			Format: envOr("LOG_FORMAT", "text"),
		},
		Rest: RestConfig{
			Port: envIntOr("REST_PORT", 8001),
			Timeouts: RestTimeouts{
				Read:       envDurationOr("REST_READ_TIMEOUT", 15*time.Second),
				ReadHeader: envDurationOr("REST_READ_HEADER_TIMEOUT", 15*time.Second),
				Write:      envDurationOr("REST_WRITE_TIMEOUT", 15*time.Second),
				Idle:       envDurationOr("REST_IDLE_TIMEOUT", 60*time.Second),
			},
		},
		GRPC: GRPCConfig{
			Port: envIntOr("GRPC_PORT", 9001),
		},
		Auth: AuthConfig{
			Tokens: AuthTokensConfig{
				Issuer: envOr("AUTH_TOKENS_ISSUER", "auth-svc"),
				UserAccess: TokenConfig{
					SecretKey: mustEnv("AUTH_TOKENS_USER_ACCESS_SECRET_KEY"),
					TTL:       envDurationOr("AUTH_TOKENS_USER_ACCESS_TTL", 720*time.Hour),
				},
				UserRefresh: RefreshTokenConfig{
					SecretKey: mustEnv("AUTH_TOKENS_USER_REFRESH_SECRET_KEY"),
					HashKey:   mustEnv("AUTH_TOKENS_USER_REFRESH_HASH_KEY"),
					TTL:       envDurationOr("AUTH_TOKENS_USER_REFRESH_TTL", 720*time.Hour),
				},
			},
			OAuth: AuthOAuthConfig{
				// Optional: the service runs fine without Google login configured,
				// it just won't be able to complete that specific flow.
				Google: GoogleOAuthConfig{
					ClientID:     envOr("AUTH_OAUTH_GOOGLE_CLIENT_ID", ""),
					ClientSecret: envOr("AUTH_OAUTH_GOOGLE_CLIENT_SECRET", ""),
					RedirectURL:  envOr("AUTH_OAUTH_GOOGLE_REDIRECT_URL", ""),
				},
			},
			PassBcryptCost: envIntOr("AUTH_PASS_BCRYPT_COST", 11),
		},
		Database: DatabaseConfig{
			SQL: SQLConfig{
				URL: mustEnv("DATABASE_SQL_URL"),
			},
			Redis: RedisConfig{
				Addr:     mustEnv("REDIS_ADDR"),
				Password: envOr("REDIS_PASSWORD", ""),
				DB:       envIntOr("DATABASE_REDIS_DB", 0),
				TTL: RedisTTLConfig{
					User:     envDurationOr("DATABASE_REDIS_TTL_USER", 5*time.Minute),
					Email:    envDurationOr("DATABASE_REDIS_TTL_EMAIL", 5*time.Minute),
					Password: envDurationOr("DATABASE_REDIS_TTL_PASSWORD", 5*time.Minute),
					Session:  envDurationOr("DATABASE_REDIS_TTL_SESSION", 5*time.Minute),
				},
			},
		},
		S3: S3Config{
			Aws: S3AwsConfig{
				Region:          mustEnv("S3_AWS_REGION"),
				BucketName:      mustEnv("S3_AWS_BUCKET_NAME"),
				AccessKeyID:     mustEnv("S3_AWS_ACCESS_KEY_ID"),
				SecretAccessKey: mustEnv("S3_AWS_SECRET_ACCESS_KEY"),
				SessionToken:    envOr("S3_AWS_SESSION_TOKEN", ""),
				BaseURL:         mustEnv("S3_AWS_BASE_URL"),
			},
			Media: S3MediaConfig{
				Link: S3MediaLinkConfig{
					TTL: envDurationOr("S3_MEDIA_LINK_TTL", 24*time.Hour),
				},
				Resources: S3MediaResourcesConfig{
					User: S3MediaUserConfig{
						Avatar: awsx.ImageValidator{
							AllowedFormats: envListOr(
								"S3_MEDIA_USER_AVATAR_ALLOWED_FORMATS", []string{"jpeg", "jpg", "png", "gif"},
							),
							MaxWidth:       envIntOr("S3_MEDIA_USER_AVATAR_MAX_WIDTH", 512),
							MinWidth:       envIntOr("S3_MEDIA_USER_AVATAR_MIN_WIDTH", 512),
							MaxHeight:      envIntOr("S3_MEDIA_USER_AVATAR_MAX_HEIGHT", 512),
							MinHeight:      envIntOr("S3_MEDIA_USER_AVATAR_MIN_HEIGHT", 512),
							ContentSizeMax: envInt64Or("S3_MEDIA_USER_AVATAR_CONTENT_SIZE_MAX", 5242880),
						},
					},
				},
			},
		},
		Kafka: KafkaConfig{
			// Not read by this service yet — Debezium publishes the outbox
			// table to Kafka independently. Identity is only used as the
			// outbox rows' producer label.
			Brokers:  envList("KAFKA_BROKERS"),
			Identity: envOr("KAFKA_IDENTITY", "auth-svc"),
		},
		OTEL: OTELConfig{
			CollectorEndpoint: envOr("OTEL_COLLECTOR_ENDPOINT", ""),
			SamplingRatio:     envFloatOr("OTEL_SAMPLING_RATIO", 1.0),
		},
	}
}

func mustEnv(key string) string {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		panic(fmt.Errorf("%s env var is not set", key))
	}
	return v
}

func envOr(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func envList(key string) []string {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return nil
	}

	return splitList(v)
}

func envListOr(key string, def []string) []string {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def
	}
	return splitList(v)
}

func splitList(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func envIntOr(key string, def int) int {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def
	}

	n, err := strconv.Atoi(v)
	if err != nil {
		panic(fmt.Errorf("invalid int value for %s: %w", key, err))
	}
	return n
}

func envInt64Or(key string, def int64) int64 {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def
	}

	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		panic(fmt.Errorf("invalid int value for %s: %w", key, err))
	}
	return n
}

func envFloatOr(key string, def float64) float64 {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def
	}

	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		panic(fmt.Errorf("invalid float value for %s: %w", key, err))
	}
	return f
}

func envDurationOr(key string, def time.Duration) time.Duration {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def
	}

	d, err := time.ParseDuration(v)
	if err != nil {
		panic(fmt.Errorf("invalid duration value for %s: %w", key, err))
	}
	return d
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
