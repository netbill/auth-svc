package account

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/core/models"
)

func (s Service) Refresh(ctx context.Context, oldRefreshToken string) (models.TokensPair, error) {
	tokenData, err := s.jwt.ParseRefreshClaims(oldRefreshToken)
	if err != nil {
		return models.TokensPair{}, errx.ErrorInternal.Raise(
			fmt.Errorf("failed to decrypt refresh token claims, cause: %w", err),
		)
	}

	accountID, err := uuid.Parse(tokenData.Subject)
	if err != nil {
		return models.TokensPair{}, errx.ErrorInternal.Raise(
			fmt.Errorf("failed to parse account id from token claims, cause: %w", err),
		)
	}

	account, err := s.GetAccountByID(ctx, accountID)
	if err != nil {
		return models.TokensPair{}, err
	}

	if err = account.CanInteract(); err != nil {
		return models.TokensPair{}, err
	}

	token, err := s.repo.GetSessionToken(ctx, tokenData.SessionID)
	if err != nil {
		return models.TokensPair{}, errx.ErrorInternal.Raise(
			fmt.Errorf("failed to get session with id: %s for account %s, cause: %w", tokenData.SessionID, accountID, err),
		)
	}
	if token == "" {
		return models.TokensPair{}, errx.ErrorSessionNotFound.Raise(
			fmt.Errorf("failed to find session with id %s for account %s, cause: %w", tokenData.SessionID, accountID, err),
		)
	}

	refresh, err := s.jwt.DecryptRefresh(token)
	if err != nil {
		return models.TokensPair{}, errx.ErrorInternal.Raise(
			fmt.Errorf("failed to generate refresh token for account %s, cause: %w", accountID, err),
		)
	}
	if refresh != oldRefreshToken {
		return models.TokensPair{}, errx.ErrorSessionTokenMismatch.Raise(
			fmt.Errorf(
				"refresh token does not match for session %s and account %s, cause: %w",
				tokenData.SessionID, accountID, err,
			),
		)
	}

	refresh, err = s.jwt.GenerateRefresh(account, tokenData.SessionID)
	if err != nil {
		return models.TokensPair{}, errx.ErrorInternal.Raise(
			fmt.Errorf("failed to generate refresh token for account %s, cause: %w", accountID, err),
		)
	}

	refreshCrypto, err := s.jwt.EncryptRefresh(refresh)
	if err != nil {
		return models.TokensPair{}, errx.ErrorInternal.Raise(
			fmt.Errorf("failed to encrypt refresh token for account %s, cause: %w", accountID, err),
		)
	}

	access, err := s.jwt.GenerateAccess(account, tokenData.SessionID)
	if err != nil {
		return models.TokensPair{}, errx.ErrorInternal.Raise(
			fmt.Errorf("failed to generate access token for account %s, cause: %w", accountID, err),
		)
	}

	_, err = s.repo.UpdateSessionToken(ctx, tokenData.SessionID, refreshCrypto)
	if err != nil {
		return models.TokensPair{}, errx.ErrorInternal.Raise(
			fmt.Errorf("failed to save refresh token for account %s, cause: %w", accountID, err),
		)
	}

	return models.TokensPair{
		SessionID: tokenData.SessionID,
		Refresh:   refresh,
		Access:    access,
	}, nil
}
