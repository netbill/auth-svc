package session

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/restkit/tokens"
)

type Service struct {
	auth auth

	accountRepo  accountRepo
	emailRepo    emailRepo
	passwordRepo passwordRepo
	sessionRepo  sessionRepo
	tx           transaction

	passwordCache passwordCache
	accountCache  accountCache
	sessionsCache sessionsCache

	passManager  passwordManager
	tokenManager tokenManager

	log *slog.Logger
}

type ServiceDeps struct {
	Auth          auth
	AccountRepo   accountRepo
	EmailRepo     emailRepo
	PasswordRepo  passwordRepo
	SessionRepo   sessionRepo
	Tx            transaction
	PasswordCache passwordCache
	AccountCache  accountCache
	SessionsCache sessionsCache
	PassManager   passwordManager
	TokenManager  tokenManager
	Log           *slog.Logger
}

func New(deps ServiceDeps) *Service {
	return &Service{
		auth:          deps.Auth,
		accountRepo:   deps.AccountRepo,
		emailRepo:     deps.EmailRepo,
		passwordRepo:  deps.PasswordRepo,
		sessionRepo:   deps.SessionRepo,
		tx:            deps.Tx,
		passwordCache: deps.PasswordCache,
		accountCache:  deps.AccountCache,
		sessionsCache: deps.SessionsCache,
		passManager:   deps.PassManager,
		tokenManager:  deps.TokenManager,
		log:           deps.Log,
	}
}

type passwordManager interface {
	CheckMatch(password, hash string) error
}

type tokenManager interface {
	ParseAccountAuthAccess(token string) (tokens.AccountAuthClaims, error)
	ParseAccountAuthRefresh(token string) (tokens.AccountAuthClaims, error)

	HashRefresh(token string) (string, error)

	GenerateAccess(account models.Account, sessionID uuid.UUID) (string, error)
	GenerateRefresh(account models.Account, sessionID uuid.UUID) (string, error)
}

func (s *Service) GetMySession(
	ctx context.Context,
	actor models.AccountActor,
	sessionID uuid.UUID,
) (models.Session, error) {
	session, err := s.sessionsCache.GetByID(ctx, sessionID)
	switch {
	case err == nil:
		if session.AccountID != actor.ID {
			return models.Session{}, errx.ErrorSessionNotFound.Raise(
				fmt.Errorf("session %s does not belong to account %s", sessionID, actor.ID),
			)
		}
		if session.DeletedAt != nil {
			return models.Session{}, errx.ErrorSessionNotFound.Raise(
				fmt.Errorf("session %s is deleted", sessionID),
			)
		}

		return session, nil
	case errors.Is(err, errx.ErrCacheMiss):
		s.log.Debug("session cache miss", "session_id", sessionID)
	default:
		s.log.Error("failed to get session from cache", "error", err)
	}

	session, err = s.sessionRepo.GetForAccount(ctx, actor.ID, sessionID)
	if err != nil {
		return models.Session{}, err
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()

		if err := s.sessionsCache.Set(ctx, session); err != nil {
			s.log.Error("failed to set session cache", "error", err)
		}
	}()

	return session, nil
}

func (s *Service) Refresh(
	ctx context.Context,
	oldRefreshToken string,
) (models.TokensPair, error) {
	claims, err := s.tokenManager.ParseAccountAuthRefresh(oldRefreshToken)
	if err != nil {
		return models.TokensPair{}, errx.ErrorSessionExpired.Raise(err)
	}

	storedHash, err := s.sessionRepo.GetToken(ctx, claims.SessionID)
	if err != nil {
		return models.TokensPair{}, err
	}

	tokenHash, err := s.tokenManager.HashRefresh(oldRefreshToken)
	if err != nil {
		return models.TokensPair{}, err
	}

	if storedHash != tokenHash {
		return models.TokensPair{}, errx.ErrorSessionTokenMismatch.Raise(fmt.Errorf("refresh token hash mismatch for session %v", claims.SessionID))
	}

	accountID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return models.TokensPair{}, err
	}

	account, err := s.accountCache.GetByID(ctx, accountID)
	switch {
	case errors.Is(err, errx.ErrCacheMiss):
		s.log.Debug("account cache miss", "account_id", accountID)
	case err != nil:
		s.log.Error("failed to get account from cache", "error", err)
	}

	if err != nil {
		account, err = s.accountRepo.GetByID(ctx, accountID)
		if err != nil {
			return models.TokensPair{}, err
		}
	}

	newRefreshToken, err := s.tokenManager.GenerateRefresh(account, claims.SessionID)
	if err != nil {
		return models.TokensPair{}, err
	}

	newHash, err := s.tokenManager.HashRefresh(newRefreshToken)
	if err != nil {
		return models.TokensPair{}, err
	}

	var session models.Session
	if err = s.tx.Transaction(ctx, func(ctx context.Context) error {
		session, err = s.sessionRepo.UpdateToken(ctx, claims.SessionID, newHash)
		return err
	}); err != nil {
		return models.TokensPair{}, err
	}

	accessToken, err := s.tokenManager.GenerateAccess(account, session.ID)
	if err != nil {
		return models.TokensPair{}, err
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()

		if err := s.sessionsCache.Set(ctx, session); err != nil {
			s.log.Error("failed to set session cache", "error", err)
		}
		if err := s.accountCache.Set(ctx, account); err != nil {
			s.log.Error("failed to set account cache", "error", err)
		}
	}()

	return models.TokensPair{
		SessionID: session.ID,
		Refresh:   newRefreshToken,
		Access:    accessToken,
	}, nil
}

func (s *Service) Logout(
	ctx context.Context,
	actor models.AccountActor,
) error {
	if err := s.sessionRepo.Delete(ctx, actor.SessionID); err != nil {
		return err
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()

		if err := s.sessionsCache.DeleteByID(ctx, actor.SessionID); err != nil {
			s.log.Error("failed to delete session cache", "error", err)
		}
	}()

	return nil
}

func (s *Service) DeleteMySession(
	ctx context.Context,
	actor models.AccountActor,
	sessionID uuid.UUID,
) error {
	if _, _, err := s.auth.ValidateSession(ctx, actor); err != nil {
		return err
	}

	if err := s.tx.Transaction(ctx, func(ctx context.Context) error {
		return s.sessionRepo.DeleteOneForAccount(ctx, actor.ID, sessionID)
	}); err != nil {
		return err
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()

		if err := s.sessionsCache.DeleteByID(ctx, sessionID); err != nil {
			s.log.Error("failed to delete session cache", "error", err)
		}
	}()

	return nil
}

func (s *Service) DeleteMySessions(
	ctx context.Context,
	actor models.AccountActor,
) error {
	if _, _, err := s.auth.ValidateSession(ctx, actor); err != nil {
		return err
	}

	var sessionIDs []uuid.UUID
	var err error
	if err = s.tx.Transaction(ctx, func(ctx context.Context) error {
		sessionIDs, err = s.sessionRepo.DeleteManyForAccount(ctx, actor.ID)
		return err
	}); err != nil {
		return err
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()

		for _, id := range sessionIDs {
			if err := s.sessionsCache.DeleteByID(ctx, id); err != nil {
				s.log.Error("failed to delete session cache", "session_id", id, "error", err)
			}
		}
	}()

	return nil
}
