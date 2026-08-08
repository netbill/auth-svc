package session

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/restkit/tokens"
)

//go:generate mockery --name=auth --inpackage
type auth interface {
	ValidateSession(ctx context.Context, actor models.UserActor) (models.User, models.Session, error)
}

//go:generate mockery --name=passwordManager --inpackage
type passwordManager interface {
	CheckMatch(password, hash string) error
}

//go:generate mockery --name=tokenManager --inpackage
type tokenManager interface {
	ParseUserAuthAccess(token string) (tokens.AccountAuthClaims, error)
	ParseUserAuthRefresh(token string) (tokens.AccountAuthClaims, error)

	HashRefresh(token string) (string, error)

	GenerateAccess(user models.User, sessionID uuid.UUID) (string, error)
	GenerateRefresh(user models.User, sessionID uuid.UUID) (string, error)
}

//go:generate mockery --name=bus --inpackage
type bus interface {
	PublishQRToken(ctx context.Context, key string, payload []byte) error
}

type Service struct {
	auth auth

	userRepo     userRepo
	emailRepo    emailRepo
	passwordRepo passwordRepo
	sessionRepo  sessionRepo
	tx           transaction

	passwordCache passwordCache
	userCache     userCache
	sessionsCache sessionsCache

	qrRepo qrRepo
	bus    bus

	passManager  passwordManager
	tokenManager tokenManager
}

type ServiceDeps struct {
	Auth auth

	UserRepo     userRepo
	EmailRepo    emailRepo
	PasswordRepo passwordRepo
	SessionRepo  sessionRepo

	Tx transaction

	PasswordCache passwordCache
	UserCache     userCache
	SessionsCache sessionsCache

	PassManager  passwordManager
	TokenManager tokenManager
	QRStore      qrRepo
	Bus          bus
}

func New(deps ServiceDeps) *Service {
	return &Service{
		auth:          deps.Auth,
		userRepo:      deps.UserRepo,
		emailRepo:     deps.EmailRepo,
		passwordRepo:  deps.PasswordRepo,
		sessionRepo:   deps.SessionRepo,
		tx:            deps.Tx,
		passwordCache: deps.PasswordCache,
		userCache:     deps.UserCache,
		sessionsCache: deps.SessionsCache,
		passManager:   deps.PassManager,
		tokenManager:  deps.TokenManager,
		qrRepo:        deps.QRStore,
		bus:           deps.Bus,
	}
}

func (s *Service) GetMySession(
	ctx context.Context,
	actor models.UserActor,
	sessionID uuid.UUID,
) (models.Session, error) {
	if session, err := s.sessionsCache.Get(ctx, sessionID); err == nil {
		if session.UserID != actor.ID {
			return models.Session{}, errx.ErrorSessionNotFound.Raise(
				fmt.Errorf("session %s does not belong to user %s", sessionID, actor.ID),
			)
		}
		if session.DeletedAt != nil {
			return models.Session{}, errx.ErrorSessionNotFound.Raise(
				fmt.Errorf("session %s is deleted", sessionID),
			)
		}
		return session, nil
	}

	session, err := s.sessionRepo.GetForUser(ctx, actor.ID, sessionID)
	if err != nil {
		return models.Session{}, err
	}

	go s.sessionsCache.Set(context.WithoutCancel(ctx), session)

	return session, nil
}

func (s *Service) Refresh(
	ctx context.Context,
	oldRefreshToken string,
) (models.TokensPair, error) {
	claims, err := s.tokenManager.ParseUserAuthRefresh(oldRefreshToken)
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
		return models.TokensPair{}, errx.ErrorSessionTokenMismatch.Raise(
			fmt.Errorf("refresh token hash mismatch for session %v", claims.SessionID),
		)
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return models.TokensPair{}, err
	}

	user, err := s.userCache.Get(ctx, userID)
	if err != nil {
		user, err = s.userRepo.GetByID(ctx, userID)
		if err != nil {
			return models.TokensPair{}, err
		}
	}

	newRefreshToken, err := s.tokenManager.GenerateRefresh(user, claims.SessionID)
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

	accessToken, err := s.tokenManager.GenerateAccess(user, session.ID)
	if err != nil {
		return models.TokensPair{}, err
	}

	detached := context.WithoutCancel(ctx)
	go s.sessionsCache.Set(detached, session)
	go s.userCache.Set(detached, user)

	return models.TokensPair{
		SessionID: session.ID,
		Refresh:   newRefreshToken,
		Access:    accessToken,
	}, nil
}

func (s *Service) Logout(
	ctx context.Context,
	actor models.UserActor,
) error {
	if err := s.sessionRepo.Delete(ctx, actor.SessionID); err != nil {
		return err
	}

	go s.sessionsCache.Delete(context.WithoutCancel(ctx), actor.SessionID)

	return nil
}

func (s *Service) DeleteMySession(
	ctx context.Context,
	actor models.UserActor,
	sessionID uuid.UUID,
) error {
	if _, _, err := s.auth.ValidateSession(ctx, actor); err != nil {
		return err
	}

	if err := s.sessionRepo.DeleteOneForUser(ctx, actor.ID, sessionID); err != nil {
		return err
	}

	go s.sessionsCache.Delete(context.WithoutCancel(ctx), sessionID)

	return nil
}

func (s *Service) DeleteMySessions(
	ctx context.Context,
	actor models.UserActor,
) error {
	if _, _, err := s.auth.ValidateSession(ctx, actor); err != nil {
		return err
	}

	sessionIDs, err := s.sessionRepo.DeleteManyForUser(ctx, actor.ID)
	if err != nil {
		return err
	}

	detached := context.WithoutCancel(ctx)
	for _, id := range sessionIDs {
		id := id
		go s.sessionsCache.Delete(detached, id)
	}

	return nil
}
