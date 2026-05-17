package session

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/restkit/tokens"
)

type auth interface {
	ValidateSession(ctx context.Context, actor models.AccountActor) (models.Account, models.Session, error)
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

type bus interface {
	PublishQRToken(ctx context.Context, key string, payload []byte) error
}

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

	qrRepo qrRepo
	bus    bus

	passManager  passwordManager
	tokenManager tokenManager
}

type ServiceDeps struct {
	Auth auth

	AccountRepo  accountRepo
	EmailRepo    emailRepo
	PasswordRepo passwordRepo
	SessionRepo  sessionRepo

	Tx transaction

	PasswordCache passwordCache
	AccountCache  accountCache
	SessionsCache sessionsCache

	PassManager  passwordManager
	TokenManager tokenManager
	QRStore      qrRepo
	Bus          bus
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
		qrRepo:        deps.QRStore,
		bus:           deps.Bus,
	}
}

func (s *Service) GetMySession(
	ctx context.Context,
	actor models.AccountActor,
	sessionID uuid.UUID,
) (models.Session, error) {
	if session, ok := s.sessionsCache.Get(ctx, sessionID); ok {
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
	}

	session, err := s.sessionRepo.GetForAccount(ctx, actor.ID, sessionID)
	if err != nil {
		return models.Session{}, err
	}

	s.sessionsCache.Set(ctx, session)

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

	account, ok := s.accountCache.Get(ctx, accountID)
	if !ok {
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

	s.sessionsCache.Set(ctx, session)
	s.accountCache.Set(ctx, account)

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

	s.sessionsCache.Delete(ctx, actor.SessionID)

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

	if err := s.sessionRepo.DeleteOneForAccount(ctx, actor.ID, sessionID); err != nil {
		return err
	}

	s.sessionsCache.Delete(ctx, sessionID)

	return nil
}

func (s *Service) DeleteMySessions(
	ctx context.Context,
	actor models.AccountActor,
) error {
	if _, _, err := s.auth.ValidateSession(ctx, actor); err != nil {
		return err
	}

	sessionIDs, err := s.sessionRepo.DeleteManyForAccount(ctx, actor.ID)
	if err != nil {
		return err
	}

	for _, id := range sessionIDs {
		s.sessionsCache.Delete(ctx, id)
	}

	return nil
}
