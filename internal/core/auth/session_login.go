package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/models"
)

func (s *Service) LoginByEmail(ctx context.Context, email, password string) (models.TokensPair, error) {
	account, err := s.account.GetByEmail(ctx, email)
	if err != nil {
		return models.TokensPair{}, err
	}

	passData, err := s.password.GetByID(ctx, account.ID)
	if err != nil {
		return models.TokensPair{}, err
	}

	if err = s.passManager.CheckPasswordMatch(password, passData.Hash); err != nil {
		return models.TokensPair{}, err
	}

	return s.createSession(ctx, account)
}

func (s *Service) LoginByUsername(ctx context.Context, username, password string) (models.TokensPair, error) {
	account, err := s.account.GetByUsername(ctx, username)
	if err != nil {
		return models.TokensPair{}, err
	}

	passData, err := s.password.GetByID(ctx, account.ID)
	if err != nil {
		return models.TokensPair{}, err
	}

	if err = s.passManager.CheckPasswordMatch(password, passData.Hash); err != nil {
		return models.TokensPair{}, err
	}

	return s.createSession(ctx, account)
}

func (s *Service) LoginByGoogle(ctx context.Context, email string) (models.TokensPair, error) {
	account, err := s.account.GetByEmail(ctx, email)
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

	access, err := s.token.GenerateAccess(account, sessionID)
	if err != nil {
		return models.TokensPair{}, err
	}

	refresh, err := s.token.GenerateRefresh(account, sessionID)
	if err != nil {
		return models.TokensPair{}, err
	}

	refreshHash, err := s.token.HashRefresh(refresh)
	if err != nil {
		return models.TokensPair{}, err
	}

	_, err = s.session.Create(ctx, sessionID, account.ID, refreshHash)
	if err != nil {
		return models.TokensPair{}, err
	}

	return models.TokensPair{
		SessionID: sessionID,
		Refresh:   refresh,
		Access:    access,
	}, nil
}

func (s *Service) createTokensPair(
	sessionID uuid.UUID,
	account models.Account,
) (models.TokensPair, error) {
	access, err := s.token.GenerateAccess(account, sessionID)
	if err != nil {
		return models.TokensPair{}, err
	}

	refresh, err := s.token.GenerateRefresh(account, sessionID)
	if err != nil {
		return models.TokensPair{}, err
	}

	return models.TokensPair{
		SessionID: sessionID,
		Refresh:   refresh,
		Access:    access,
	}, nil
}
