package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AuthServiceSuite struct {
	suite.Suite

	accountRepo *mockAccountRepo
	sessionRepo *mockSessionRepo

	svc *Service
}

func (s *AuthServiceSuite) SetupTest() {
	s.accountRepo = newMockAccountRepo(s.T())
	s.sessionRepo = newMockSessionRepo(s.T())

	s.svc = New(ServiceDeps{
		AccountRepo: s.accountRepo,
		SessionRepo: s.sessionRepo,
	})
}

func TestAuthService(t *testing.T) {
	suite.Run(t, new(AuthServiceSuite))
}

// ─── ValidateSession ─────────────────────────────────────────────────────────

func (s *AuthServiceSuite) TestValidateSession_HappyPath() {
	accountID := uuid.New()
	sessionID := uuid.New()
	account := models.Account{ID: accountID}
	session := models.Session{ID: sessionID}
	actor := models.AccountActor{ID: accountID, SessionID: sessionID}

	s.accountRepo.On("GetByID", mock.Anything, accountID).Return(account, nil)
	s.sessionRepo.On("GetByID", mock.Anything, sessionID).Return(session, nil)

	gotAccount, gotSession, err := s.svc.ValidateSession(context.Background(), actor)

	require.NoError(s.T(), err)
	assert.Equal(s.T(), account, gotAccount)
	assert.Equal(s.T(), session, gotSession)
}

func (s *AuthServiceSuite) TestValidateSession_AccountNotFound() {
	accountID := uuid.New()
	sessionID := uuid.New()
	actor := models.AccountActor{ID: accountID, SessionID: sessionID}

	s.accountRepo.On("GetByID", mock.Anything, accountID).Return(models.Account{}, errx.ErrorAccountNotFound)

	_, _, err := s.svc.ValidateSession(context.Background(), actor)

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, errx.ErrorAccountInvalidSession)
}

func (s *AuthServiceSuite) TestValidateSession_AccountRepoError() {
	accountID := uuid.New()
	sessionID := uuid.New()
	repoErr := errors.New("db connection lost")
	actor := models.AccountActor{ID: accountID, SessionID: sessionID}

	s.accountRepo.On("GetByID", mock.Anything, accountID).Return(models.Account{}, repoErr)

	_, _, err := s.svc.ValidateSession(context.Background(), actor)

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, repoErr)
}

func (s *AuthServiceSuite) TestValidateSession_SessionNotFound() {
	accountID := uuid.New()
	sessionID := uuid.New()
	account := models.Account{ID: accountID}
	actor := models.AccountActor{ID: accountID, SessionID: sessionID}

	s.accountRepo.On("GetByID", mock.Anything, accountID).Return(account, nil)
	s.sessionRepo.On("GetByID", mock.Anything, sessionID).Return(models.Session{}, errx.ErrorSessionNotFound)

	_, _, err := s.svc.ValidateSession(context.Background(), actor)

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, errx.ErrorAccountInvalidSession)
}

func (s *AuthServiceSuite) TestValidateSession_SessionRepoError() {
	accountID := uuid.New()
	sessionID := uuid.New()
	account := models.Account{ID: accountID}
	repoErr := errors.New("db connection lost")
	actor := models.AccountActor{ID: accountID, SessionID: sessionID}

	s.accountRepo.On("GetByID", mock.Anything, accountID).Return(account, nil)
	s.sessionRepo.On("GetByID", mock.Anything, sessionID).Return(models.Session{}, repoErr)

	_, _, err := s.svc.ValidateSession(context.Background(), actor)

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, repoErr)
}
