package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/restkit/pagi"
)

func (s *Service) validateActorSession(
	ctx context.Context,
	actor models.AccountActor,
) (models.Account, models.Session, error) {
	account, err := s.account.GetByID(ctx, actor.ID)
	if errors.Is(err, errx.ErrorAccountNotFound) {
		buried, err := s.tombstone.AccountIsBuried(ctx, actor.ID)
		if err != nil {
			return models.Account{}, models.Session{}, err
		}
		if buried {
			return models.Account{}, models.Session{}, errx.ErrorAccountInvalidSession.Raise(
				fmt.Errorf("account with id %s is buried", actor.ID),
			)
		}

		return models.Account{}, models.Session{}, errx.ErrorAccountInvalidSession.Raise(
			fmt.Errorf("account with id %s not found", actor.ID),
		)
	}
	if err != nil {
		return models.Account{}, models.Session{}, err
	}

	session, err := s.session.GetByID(ctx, actor.SessionID)
	if errors.Is(err, errx.ErrorSessionNotFound) {
		buried, err := s.tombstone.SessionIsBuried(ctx, actor.SessionID)
		if err != nil {
			return models.Account{}, models.Session{}, err
		}
		if buried {
			return models.Account{}, models.Session{}, errx.ErrorAccountInvalidSession.Raise(
				fmt.Errorf("session with id %s is buried", actor.SessionID),
			)
		}

		return models.Account{}, models.Session{}, errx.ErrorAccountInvalidSession.Raise(
			fmt.Errorf("session with id %s not found", actor.SessionID),
		)
	}
	if err != nil {
		return models.Account{}, models.Session{}, err
	}

	return account, session, nil
}

func (s *Service) GetMySession(
	ctx context.Context,
	actor models.AccountActor,
	sessionID uuid.UUID,
) (models.Session, error) {
	_, _, err := s.validateActorSession(ctx, actor)
	if err != nil {
		return models.Session{}, err
	}

	session, err := s.session.GetForAccount(ctx, actor.ID, sessionID)
	if err != nil {
		return models.Session{}, err
	}

	return session, nil
}

func (s *Service) GetMySessions(
	ctx context.Context,
	actor models.AccountActor,
	limit, offset uint,
) (pagi.Page[[]models.Session], error) {
	_, _, err := s.validateActorSession(ctx, actor)
	if err != nil {
		return pagi.Page[[]models.Session]{}, err
	}

	sessions, err := s.session.GetListForAccount(ctx, actor.ID, limit, offset)
	if err != nil {
		return pagi.Page[[]models.Session]{}, err
	}

	return sessions, nil
}

func (s *Service) Refresh(ctx context.Context, oldRefreshToken string) (models.TokensPair, error) {
	tokenData, err := s.token.ParseAccountAuthRefresh(oldRefreshToken)
	if err != nil {
		return models.TokensPair{}, err
	}

	account, err := s.account.GetByID(ctx, tokenData.GetAccountID())
	if err != nil {
		return models.TokensPair{}, err
	}

	storedHash, err := s.session.GetToken(ctx, tokenData.SessionID)
	if err != nil {
		return models.TokensPair{}, err
	}

	incomingHash, err := s.token.HashRefresh(oldRefreshToken)
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

	refresh, err := s.token.GenerateRefresh(account, tokenData.SessionID)
	if err != nil {
		return models.TokensPair{}, err
	}

	refreshNewHash, err := s.token.HashRefresh(refresh)
	if err != nil {
		return models.TokensPair{}, err
	}

	access, err := s.token.GenerateAccess(account, tokenData.SessionID)
	if err != nil {
		return models.TokensPair{}, err
	}

	_, err = s.session.UpdateToken(ctx, tokenData.SessionID, refreshNewHash)
	if err != nil {
		return models.TokensPair{}, err
	}

	return models.TokensPair{
		SessionID: tokenData.SessionID,
		Refresh:   refresh,
		Access:    access,
	}, nil
}

func (s *Service) Logout(
	ctx context.Context,
	actor models.AccountActor,
) error {
	return s.tx.Transaction(ctx, func(ctx context.Context) error {
		if err := s.tombstone.BurySession(ctx, actor.SessionID); err != nil {
			return err
		}

		return s.session.DeleteOneForAccount(ctx, actor.ID, actor.SessionID)
	})
}

func (s *Service) DeleteMySession(
	ctx context.Context,
	actor models.AccountActor,
	sessionID uuid.UUID,
) error {
	_, _, err := s.validateActorSession(ctx, actor)
	if err != nil {
		return err
	}

	return s.tx.Transaction(ctx, func(ctx context.Context) error {
		if err := s.tombstone.BurySession(ctx, sessionID); err != nil {
			return err
		}

		return s.session.DeleteOneForAccount(ctx, actor.ID, sessionID)
	})
}

func (s *Service) DeleteMySessions(ctx context.Context, actor models.AccountActor) error {
	_, _, err := s.validateActorSession(ctx, actor)
	if err != nil {
		return err
	}

	return s.tx.Transaction(ctx, func(ctx context.Context) error {
		if err := s.tombstone.BuryAccountSessions(ctx, actor.ID); err != nil {
			return err
		}

		return s.session.DeleteManyForAccount(ctx, actor.ID)
	})
}
