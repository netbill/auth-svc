package auth

import (
	"context"
	"fmt"

	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/core/models"
)

func (m *Module) Refresh(ctx context.Context, oldRefreshToken string) (models.TokensPair, error) {
	tokenData, err := m.token.ParseAccountAuthRefresh(oldRefreshToken)
	if err != nil {
		return models.TokensPair{}, err
	}

	account, err := m.repo.GetAccountByID(ctx, tokenData.GetAccountID())
	if err != nil {
		return models.TokensPair{}, err
	}

	storedHash, err := m.repo.GetSessionToken(ctx, tokenData.SessionID)
	if err != nil {
		return models.TokensPair{}, err
	}

	incomingHash, err := m.token.HashRefresh(oldRefreshToken)
	if err != nil {
		return models.TokensPair{}, err
	}

	if incomingHash != storedHash {
		return models.TokensPair{}, errx.ErrorSessionTokenMismatch.Raise(
			fmt.Errorf(
				"refresh token does not match for session %s and account %s",
				tokenData.SessionID, tokenData.GetAccountID(),
			),
		)
	}

	refresh, err := m.token.GenerateRefresh(account, tokenData.SessionID)
	if err != nil {
		return models.TokensPair{}, err
	}

	refreshNewHash, err := m.token.HashRefresh(refresh)
	if err != nil {
		return models.TokensPair{}, err
	}

	access, err := m.token.GenerateAccess(account, tokenData.SessionID)
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
