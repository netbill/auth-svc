package domain_test

import (
	"context"
	"database/sql"
	"log"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/umisto/sso-svc/cmd/migrations"
	"github.com/umisto/sso-svc/internal"
	"github.com/umisto/sso-svc/internal/domain/models"
	"github.com/umisto/sso-svc/internal/repository"
	"github.com/umisto/sso-svc/internal/token"
)

// TEST DATABASE CONNECTION
const testDatabaseURL = "postgresql://postgres:postgres@localhost:7777/postgres?sslmode=disable"

type SessionSvc interface {
	Delete(ctx context.Context, sessionID uuid.UUID) error
	DeleteOneForUser(ctx context.Context, userID, sessionID uuid.UUID) error
	DeleteAllForUser(ctx context.Context, userID uuid.UUID) error

	Refresh(ctx context.Context, oldRefreshToken string) (models.TokensPair, error)

	Get(ctx context.Context, sessionID uuid.UUID) (models.Session, error)
	GetForUser(ctx context.Context, userID, sessionID uuid.UUID) (models.Session, error)

	ListForUser(
		ctx context.Context,
		userID uuid.UUID,
		page uint64,
		size uint64,
	) (models.SessionsCollection, error)
}

type UserSvc interface {
	BlockUser(ctx context.Context, userID uuid.UUID) (models.User, error)
	UnblockUser(ctx context.Context, userID uuid.UUID) (models.User, error)

	GetByID(ctx context.Context, ID uuid.UUID) (models.User, error)
	GetByEmail(ctx context.Context, email string) (models.User, error)
}

type AuthSvc interface {
	Register(
		ctx context.Context,
		email, pass, role string,
	) (models.User, error)
	RegisterAdmin(
		ctx context.Context,
		initiatorID uuid.UUID,
		email, pass, role string,
	) (models.User, error)

	UpdatePassword(
		ctx context.Context,
		userID uuid.UUID,
		oldPassword, newPassword string,
	) error

	Login(ctx context.Context, email, password string) (models.TokensPair, error)
	LoginByGoogle(ctx context.Context, email string) (models.TokensPair, error)
	CreateSession(ctx context.Context, user models.User) (models.TokensPair, error)
}

type services struct {
	Session SessionSvc
	User    UserSvc
	Auth    AuthSvc
}

type Setup struct {
	core services

	Cfg internal.Config
}

func cleanDb(t *testing.T) {
	err := migrations.MigrateDown(testDatabaseURL)
	if err != nil {
		t.Fatalf("migrate down: %v", err)
	}
	err = migrations.MigrateUp(testDatabaseURL)
	if err != nil {
		t.Fatalf("migrate up: %v", err)
	}
}

func newSetup(t *testing.T) (Setup, error) {
	cfg := internal.Config{
		Database: internal.DatabaseConfig{
			SQL: struct {
				URL string `mapstructure:"url"`
			}{
				URL: testDatabaseURL,
			},
		},
		JWT: internal.JWTConfig{
			User: struct {
				AccessToken struct {
					SecretKey     string        `mapstructure:"secret_key"`
					TokenLifetime time.Duration `mapstructure:"token_lifetime"`
				} `mapstructure:"access_token"`
				RefreshToken struct {
					SecretKey     string        `mapstructure:"secret_key"`
					EncryptionKey string        `mapstructure:"encryption_key"`
					TokenLifetime time.Duration `mapstructure:"token_lifetime"`
				} `mapstructure:"refresh_token"`
			}{
				AccessToken: struct {
					SecretKey     string        `mapstructure:"secret_key"`
					TokenLifetime time.Duration `mapstructure:"token_lifetime"`
				}{
					SecretKey:     "UnG06MAU2i1Mvqf8", //example
					TokenLifetime: time.Minute * 15,
				},
				RefreshToken: struct {
					SecretKey     string        `mapstructure:"secret_key"`
					EncryptionKey string        `mapstructure:"encryption_key"`
					TokenLifetime time.Duration `mapstructure:"token_lifetime"`
				}{
					SecretKey:     "6DSjhhT9KIezubpR", //example
					EncryptionKey: "Zlyh20N8uojZHFdO", //example
					TokenLifetime: time.Hour * 24 * 7,
				},
			},
		},
	}

	pg, err := sql.Open("postgres", cfg.Database.SQL.URL)
	if err != nil {
		log.Fatal("failed to connect to database", "error", err)
	}

	database := repository.New(pg)

	jwtTokenManager := token.NewManager(token.Config{
		AccessSK:   cfg.JWT.User.AccessToken.SecretKey,
		RefreshSK:  cfg.JWT.User.RefreshToken.SecretKey,
		AccessTTL:  cfg.JWT.User.AccessToken.TokenLifetime,
		RefreshTTL: cfg.JWT.User.RefreshToken.TokenLifetime,
		Iss:        cfg.Service.Name,
	})

	userSvc := user.New(database)
	sessionSvc := session.New(database, jwtTokenManager)
	authSvc := auth.New(database, jwtTokenManager)

	return Setup{
		core: services{
			User:    userSvc,
			Session: sessionSvc,
			Auth:    authSvc,
		},
		Cfg: cfg,
	}, nil
}
