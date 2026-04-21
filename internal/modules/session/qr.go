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

const (
	qrTokenTTL     = 5 * time.Minute
	qrConfirmedTTL = 30 * time.Second
)

type qrRepo interface {
	Set(ctx context.Context, token string, status string, ttl time.Duration) error
	Get(ctx context.Context, token string) (string, error)
}

func (s *Service) CreateQRToken(ctx context.Context) (string, error) {
	token := uuid.New().String()
	if err := s.qrRepo.Set(ctx, token, "pending", qrTokenTTL); err != nil {
		return "", err
	}

	return token, nil
}

func (s *Service) ConfirmQRToken(
	ctx context.Context,
	actor models.AccountActor,
	qrToken string,
) (models.TokensPair, error) {
	status, err := s.qrRepo.Get(ctx, qrToken)
	switch {
	case errors.Is(err, errx.ErrCacheMiss):
		return models.TokensPair{}, errx.ErrorQRTokenNotFound.Raise(fmt.Errorf("qr token %s not found or expired", qrToken))
	case err != nil:
		return models.TokensPair{}, err
	}

	if status != "pending" {
		return models.TokensPair{}, errx.ErrorQRTokenAlreadyConfirmed.Raise(fmt.Errorf("qr token %s already confirmed", qrToken))
	}

	account, err := s.accountRepo.GetByID(ctx, actor.ID)
	if err != nil {
		return models.TokensPair{}, err
	}

	pair, err := s.createSession(ctx, account)
	if err != nil {
		return models.TokensPair{}, err
	}

	if err = s.qrRepo.Set(ctx, qrToken, "confirmed", qrConfirmedTTL); err != nil {
		s.log.Error("failed to update qr token status", "error", err)
	}

	//payload, err := json.Marshal(pair)
	//if err != nil {
	//	return models.TokensPair{}, err
	//}
	//
	//if err = s.qrPublisher.Publish(ctx, qrToken, payload); err != nil {
	//	s.log.Error("failed to publish qr confirm", "error", err)
	//}

	return pair, nil
}
