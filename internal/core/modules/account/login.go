package account

import (
	"context"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/models"
)

func (m *Module) LoginByEmail(ctx context.Context, email, password string) (models.TokensPair, error) {
	account, err := m.GetAccountByEmail(ctx, email)
	if err != nil {
		return models.TokensPair{}, err
	}

	err = m.checkAccountPassword(ctx, account.ID, password)
	if err != nil {
		return models.TokensPair{}, err
	}

	return m.createSession(ctx, account)
}

func (m *Module) LoginByGoogle(ctx context.Context, email string) (models.TokensPair, error) {
	account, err := m.GetAccountByEmail(ctx, email)
	if err != nil {
		return models.TokensPair{}, err
	}

	return m.createSession(ctx, account)
}

func (m *Module) LoginByUsername(ctx context.Context, username, password string) (models.TokensPair, error) {
	account, err := m.GetAccountByUsername(ctx, username)
	if err != nil {
		return models.TokensPair{}, err
	}

	err = m.checkAccountPassword(ctx, account.ID, password)
	if err != nil {
		return models.TokensPair{}, err
	}

	return m.createSession(ctx, account)
}

func (m *Module) checkAccountPassword(
	ctx context.Context,
	accountID uuid.UUID,
	password string,
) error {
	passData, err := m.repo.GetAccountPassword(ctx, accountID)
	if err != nil {
		return err
	}

	if err = passData.CheckPasswordMatch(password); err != nil {
		return err
	}

	return nil
}

func (m *Module) createSession(
	ctx context.Context,
	account models.Account,
) (models.TokensPair, error) {
	sessionID := uuid.New()

	pair, err := m.createTokensPair(sessionID, account)
	if err != nil {
		return models.TokensPair{}, err
	}

	refreshHash, err := m.jwt.HashRefresh(pair.Refresh)
	if err != nil {
		return models.TokensPair{}, err
	}

	_, err = m.repo.CreateSession(ctx, sessionID, account.ID, refreshHash)
	if err != nil {
		return models.TokensPair{}, err
	}

	return models.TokensPair{
		SessionID: pair.SessionID,
		Refresh:   pair.Refresh,
		Access:    pair.Access,
	}, nil
}

func (m *Module) createTokensPair(
	sessionID uuid.UUID,
	account models.Account,
) (models.TokensPair, error) {
	access, err := m.jwt.GenerateAccess(account, sessionID)
	if err != nil {
		return models.TokensPair{}, err
	}

	refresh, err := m.jwt.GenerateRefresh(account, sessionID)
	if err != nil {
		return models.TokensPair{}, err
	}

	return models.TokensPair{
		SessionID: sessionID,
		Refresh:   refresh,
		Access:    access,
	}, nil
}
