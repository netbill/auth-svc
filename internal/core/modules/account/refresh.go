package account

import (
	"context"
	"fmt"

	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/core/models"
)

func (m *Module) Refresh(ctx context.Context, oldRefreshToken string) (models.TokensPair, error) {
	tokenData, err := m.jwt.ParseAccountAuthRefreshClaims(oldRefreshToken)
	if err != nil {
		return models.TokensPair{}, err
	}

	account, err := m.GetAccountByID(ctx, tokenData.GetAccountID())
	if err != nil {
		return models.TokensPair{}, err
	}

	token, err := m.repo.GetSessionToken(ctx, tokenData.SessionID)
	if err != nil {
		return models.TokensPair{}, err
	}

	refreshHash, err := m.jwt.HashRefresh(token)
	if err != nil {
		return models.TokensPair{}, err
	}

	if refreshHash != oldRefreshToken {
		return models.TokensPair{}, errx.ErrorSessionTokenMismatch.Raise(
			fmt.Errorf(
				"refresh token does not match for session %s and account %s, cause: %w",
				tokenData.SessionID, tokenData.GetAccountID(), err,
			),
		)
	}

	refresh, err := m.jwt.GenerateRefresh(account, tokenData.SessionID)
	if err != nil {
		return models.TokensPair{}, err
	}

	refreshNewHash, err := m.jwt.HashRefresh(refresh)
	if err != nil {
		return models.TokensPair{}, err
	}

	access, err := m.jwt.GenerateAccess(account, tokenData.SessionID)
	if err != nil {
		return models.TokensPair{}, err
	}

	_, err = m.repo.UpdateSessionToken(ctx, tokenData.SessionID, refreshNewHash)
	if err != nil {
		return models.TokensPair{}, err
	}

	return models.TokensPair{
		SessionID: tokenData.SessionID,
		Refresh:   refresh,
		Access:    access,
	}, nil
}
