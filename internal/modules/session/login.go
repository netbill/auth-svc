package session

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
)

func (s *Service) LoginByEmail(
	ctx context.Context,
	email, password string,
) (models.TokensPair, error) {
	emailRecord, err := s.emailRepo.GetByEmail(ctx, email)
	if err != nil {
		return models.TokensPair{}, err
	}

	user, err := s.userRepo.GetByID(ctx, emailRecord.UserID)
	if err != nil {
		return models.TokensPair{}, err
	}

	if err = s.checkPassword(ctx, user.ID, password); err != nil {
		return models.TokensPair{}, err
	}

	return s.createSession(ctx, user)
}

func (s *Service) checkPassword(ctx context.Context, userID uuid.UUID, password string) error {
	pwd, err := s.passwordCache.Get(ctx, userID)
	if err != nil {
		pwd, err = s.passwordRepo.GetByID(ctx, userID)
		if err != nil {
			return err
		}

		go s.passwordCache.Set(context.WithoutCancel(ctx), pwd)
	}

	return s.passManager.CheckMatch(password, pwd.Hash)
}

func (s *Service) LoginByGoogle(
	ctx context.Context,
	email string,
) (models.TokensPair, error) {
	emailRecord, err := s.emailRepo.GetByEmail(ctx, email)
	if err != nil {
		return models.TokensPair{}, err
	}

	user, err := s.userRepo.GetByID(ctx, emailRecord.UserID)
	if err != nil {
		return models.TokensPair{}, err
	}

	return s.createSession(ctx, user)
}

func (s *Service) createSession(
	ctx context.Context,
	user models.User,
) (models.TokensPair, error) {
	sessionID := uuid.New()

	refreshToken, err := s.tokenManager.GenerateRefresh(user, sessionID)
	if err != nil {
		return models.TokensPair{}, err
	}

	hashToken, err := s.tokenManager.HashRefresh(refreshToken)
	if err != nil {
		return models.TokensPair{}, err
	}

	var session models.Session
	if err = s.tx.Transaction(ctx, func(ctx context.Context) error {
		session, err = s.sessionRepo.Create(ctx, sessionID, user.ID, hashToken)
		return err
	}); err != nil {
		return models.TokensPair{}, err
	}

	accessToken, err := s.tokenManager.GenerateAccess(user, session.ID)
	if err != nil {
		return models.TokensPair{}, err
	}

	detached := context.WithoutCancel(ctx)
	go s.userCache.Set(detached, user)
	go s.sessionsCache.Set(detached, session)

	return models.TokensPair{
		SessionID: session.ID,
		Refresh:   refreshToken,
		Access:    accessToken,
	}, nil
}

// QRTokenTTL is how long a freshly created QR token stays pending.
// The HTTP handler streaming the QR flow to the client uses the same value
// as its write deadline, so both sides expire in lockstep.
const QRTokenTTL = 5 * time.Minute

const qrConfirmedTTL = 30 * time.Second

//go:generate mockery --name=qrRepo --inpackage
type qrRepo interface {
	Set(ctx context.Context, token string, status string, ttl time.Duration) error
	Get(ctx context.Context, token string) (string, error)
}

func (s *Service) CreateQRToken(ctx context.Context) (string, error) {
	token := uuid.New().String()
	if err := s.qrRepo.Set(ctx, token, "pending", QRTokenTTL); err != nil {
		return "", err
	}

	return token, nil
}

func (s *Service) ConfirmQRToken(
	ctx context.Context,
	actor models.UserActor,
	qrToken string,
) (models.TokensPair, error) {
	status, err := s.qrRepo.Get(ctx, qrToken)
	switch {
	case errors.Is(err, errx.ErrorQRTokenNotFound):
		return models.TokensPair{}, err
	case err != nil:
		return models.TokensPair{}, err
	}

	if status != "pending" {
		return models.TokensPair{}, errx.ErrorQRTokenAlreadyConfirmed.Raise(
			fmt.Errorf("qr token %s already confirmed", qrToken),
		)
	}

	user, err := s.userRepo.GetByID(ctx, actor.ID)
	if err != nil {
		return models.TokensPair{}, err
	}

	pair, err := s.createSession(ctx, user)
	if err != nil {
		return models.TokensPair{}, err
	}

	if err = s.qrRepo.Set(ctx, qrToken, "confirmed", qrConfirmedTTL); err != nil {
		return models.TokensPair{}, err
	}

	return pair, nil
}

func (s *Service) PublishQRToken(ctx context.Context, key string, payload []byte) error {
	return s.bus.PublishQRToken(ctx, key, payload)
}
