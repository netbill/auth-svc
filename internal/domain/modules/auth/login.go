package auth

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/umisto/sso-svc/internal/domain/errx"
	"github.com/umisto/sso-svc/internal/domain/models"
)

func (s Service) LoginByEmail(ctx context.Context, email, password string) (models.TokensPair, error) {
	account, err := s.GetAccountByEmail(ctx, email)
	if err != nil {
		return models.TokensPair{}, err
	}

	if err = account.CanInteract(); err != nil {
		return models.TokensPair{}, err
	}

	err = s.checkAccountPassword(ctx, account.ID, password)
	if err != nil {
		return models.TokensPair{}, err
	}

	return s.createSession(ctx, account)
}

func (s Service) LoginByUsername(ctx context.Context, username, password string) (models.TokensPair, error) {
	account, err := s.GetAccountByUsername(ctx, username)
	if err != nil {
		return models.TokensPair{}, err
	}

	if err = account.CanInteract(); err != nil {
		return models.TokensPair{}, err
	}

	err = s.checkAccountPassword(ctx, account.ID, password)
	if err != nil {
		return models.TokensPair{}, err
	}

	return s.createSession(ctx, account)
}

func (s Service) LoginByGoogle(ctx context.Context, email string) (models.TokensPair, error) {
	account, err := s.GetAccountByEmail(ctx, email)
	if err != nil {
		return models.TokensPair{}, err
	}

	if err = account.CanInteract(); err != nil {
		return models.TokensPair{}, err
	}

	return s.createSession(ctx, account)
}

func (s Service) checkAccountPassword(
	ctx context.Context,
	accountID uuid.UUID,
	password string,
) error {
	passData, err := s.db.GetAccountPassword(ctx, accountID)
	if err != nil {
		return errx.ErrorInternal.Raise(
			fmt.Errorf("failed to get account password, cause: %w", err),
		)
	}

	if err = passData.CheckPasswordMatch(password); err != nil {
		return err
	}

	return nil
}

func (s Service) createSession(
	ctx context.Context,
	account models.Account,
) (models.TokensPair, error) {
	sessionID := uuid.New()

	pair, err := s.createTokensPair(sessionID, account)
	if err != nil {
		return models.TokensPair{}, err
	}

	refreshTokenCrypto, err := s.jwt.EncryptRefresh(pair.Refresh)
	if err != nil {
		return models.TokensPair{}, errx.ErrorInternal.Raise(
			fmt.Errorf("failed to encrypt refresh token for account %s, cause: %w", account.ID, err),
		)
	}

	_, err = s.db.CreateSession(ctx, sessionID, account.ID, refreshTokenCrypto)
	if err != nil {
		return models.TokensPair{}, errx.ErrorInternal.Raise(
			fmt.Errorf("failed to createSession session for account %s, cause: %w", account.ID, err),
		)
	}

	email, err := s.GetAccountEmail(ctx, account.ID)
	if err != nil {
		return models.TokensPair{}, err
	}

	err = s.event.WriteAccountLogin(ctx, account, email.Email)
	if err != nil {
		return models.TokensPair{}, errx.ErrorInternal.Raise(
			fmt.Errorf("failed to publish account login event for account %s: %w", account.ID, err),
		)
	}

	return models.TokensPair{
		SessionID: pair.SessionID,
		Refresh:   pair.Refresh,
		Access:    pair.Access,
	}, nil
}

func (s Service) createTokensPair(
	sessionID uuid.UUID,
	account models.Account,
) (models.TokensPair, error) {
	access, err := s.jwt.GenerateAccess(account, sessionID)
	if err != nil {
		return models.TokensPair{}, errx.ErrorInternal.Raise(
			fmt.Errorf("failed to generate access token for account %s, cause: %w", account.ID, err),
		)
	}

	refresh, err := s.jwt.GenerateRefresh(account, sessionID)
	if err != nil {
		return models.TokensPair{}, errx.ErrorInternal.Raise(
			fmt.Errorf("failed to generate refresh token for account %s, cause: %w", account.ID, err),
		)
	}

	return models.TokensPair{
		SessionID: sessionID,
		Refresh:   refresh,
		Access:    access,
	}, nil
}
