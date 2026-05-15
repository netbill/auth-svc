package session

import (
	"context"
	"errors"
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

	account, err := s.accountRepo.GetByID(ctx, emailRecord.AccountID)
	if err != nil {
		return models.TokensPair{}, err
	}

	if err = s.checkPassword(ctx, account.ID, password); err != nil {
		return models.TokensPair{}, err
	}

	return s.createSession(ctx, account)
}

func (s *Service) LoginByUsername(
	ctx context.Context,
	username, password string,
) (models.TokensPair, error) {
	account, err := s.accountRepo.GetByUsername(ctx, username)
	if err != nil {
		return models.TokensPair{}, err
	}

	if err = s.checkPassword(ctx, account.ID, password); err != nil {
		return models.TokensPair{}, err
	}

	return s.createSession(ctx, account)
}

func (s *Service) checkPassword(ctx context.Context, accountID uuid.UUID, password string) error {
	pwd, err := s.passwordCache.GetByID(ctx, accountID)
	switch {
	case errors.Is(err, errx.ErrCacheMiss):
		s.log.Debug("password cache miss", "account_id", accountID)
	case err != nil:
		s.log.Error("failed to get password from cache", "error", err)
	}

	if err != nil {
		pwd, err = s.passwordRepo.GetByID(ctx, accountID)
		if err != nil {
			return err
		}

		go func() {
			ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			defer cancel()

			if err := s.passwordCache.SetByID(ctx, pwd); err != nil {
				s.log.Error("failed to set password cache", "error", err)
			}
		}()
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

	account, err := s.accountRepo.GetByID(ctx, emailRecord.AccountID)
	if err != nil {
		return models.TokensPair{}, err
	}

	return s.createSession(ctx, account)
}

func (s *Service) createSession(
	ctx context.Context,
	account models.Account,
) (models.TokensPair, error) {
	sessionID := uuid.New()

	refreshToken, err := s.tokenManager.GenerateRefresh(account, sessionID)
	if err != nil {
		return models.TokensPair{}, err
	}

	hashToken, err := s.tokenManager.HashRefresh(refreshToken)
	if err != nil {
		return models.TokensPair{}, err
	}

	var session models.Session
	if err = s.tx.Transaction(ctx, func(ctx context.Context) error {
		session, err = s.sessionRepo.Create(ctx, sessionID, account.ID, hashToken)
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

		if err := s.accountCache.Set(ctx, account); err != nil {
			s.log.Error("failed to set account cache", "error", err)
		}
		if err := s.sessionsCache.Set(ctx, session); err != nil {
			s.log.Error("failed to set session cache", "error", err)
		}
	}()

	return models.TokensPair{
		SessionID: session.ID,
		Refresh:   refreshToken,
		Access:    accessToken,
	}, nil
}
