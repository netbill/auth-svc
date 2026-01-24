package account

import (
	"context"
	"fmt"

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

	account, err := s.GetAccountByID(ctx, tokenData.AccountID)
	if err != nil {
		return models.TokensPair{}, err
	}

	token, err := s.repo.GetSessionToken(ctx, tokenData.SessionID)
	if err != nil {
		return models.TokensPair{}, errx.ErrorInternal.Raise(
			fmt.Errorf("failed to get session with id: %s for account %s, cause: %w", tokenData.SessionID, tokenData.AccountID, err),
		)
	}
	if token == "" {
		return models.TokensPair{}, errx.ErrorSessionNotFound.Raise(
			fmt.Errorf("failed to find session with id %s for account %s, cause: %w", tokenData.SessionID, tokenData.AccountID, err),
		)
	}

	refreshHash, err := s.jwt.HashRefresh(token)
	if err != nil {
		return models.TokensPair{}, errx.ErrorInternal.Raise(
			fmt.Errorf("failed to generate refresh token for account %s, cause: %w", tokenData.AccountID, err),
		)
	}
	if refreshHash != oldRefreshToken {
		return models.TokensPair{}, errx.ErrorSessionTokenMismatch.Raise(
			fmt.Errorf(
				"refresh token does not match for session %s and account %s, cause: %w",
				tokenData.SessionID, tokenData.AccountID, err,
			),
		)
	}

	refresh, err := s.jwt.GenerateRefresh(account, tokenData.SessionID)
	if err != nil {
		return models.TokensPair{}, errx.ErrorInternal.Raise(
			fmt.Errorf("failed to generate refresh token for account %s, cause: %w", tokenData.AccountID, err),
		)
	}

	refreshNewHash, err := s.jwt.HashRefresh(refresh)
	if err != nil {
		return models.TokensPair{}, errx.ErrorInternal.Raise(
			fmt.Errorf("failed to encrypt refresh token for account %s, cause: %w", tokenData.AccountID, err),
		)
	}

	access, err := s.jwt.GenerateAccess(account, tokenData.SessionID)
	if err != nil {
		return models.TokensPair{}, errx.ErrorInternal.Raise(
			fmt.Errorf("failed to generate access token for account %s, cause: %w", tokenData.AccountID, err),
		)
	}

	_, err = s.repo.UpdateSessionToken(ctx, tokenData.SessionID, refreshNewHash)
	if err != nil {
		return models.TokensPair{}, errx.ErrorInternal.Raise(
			fmt.Errorf("failed to save refresh token for account %s, cause: %w", tokenData.AccountID, err),
		)
	}

	return models.TokensPair{
		SessionID: tokenData.SessionID,
		Refresh:   refresh,
		Access:    access,
	}, nil
}
