package session

import (
	"context"

	"github.com/netbill/auth-svc/internal/models"
)

func (s *Service) LoginByEmail(
	ctx context.Context,
	email, password string,
) (models.TokensPair, error) {
	return models.TokensPair{}, nil
}

func (s *Service) LoginByUsername(
	ctx context.Context,
	username, password string,
) (models.TokensPair, error) {
	return models.TokensPair{}, nil
}

func (s *Service) LoginByGoogle(
	ctx context.Context,
	email string,
) (models.TokensPair, error) {
	return models.TokensPair{}, nil
}
