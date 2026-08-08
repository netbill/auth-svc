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

	userRepo    *mockUserRepo
	sessionRepo *mockSessionRepo

	svc *Service
}

func (s *AuthServiceSuite) SetupTest() {
	s.userRepo = newMockUserRepo(s.T())
	s.sessionRepo = newMockSessionRepo(s.T())

	s.svc = New(ServiceDeps{
		UserRepo:    s.userRepo,
		SessionRepo: s.sessionRepo,
	})
}

func TestAuthService(t *testing.T) {
	suite.Run(t, new(AuthServiceSuite))
}

// ─── ValidateSession ─────────────────────────────────────────────────────────

func (s *AuthServiceSuite) TestValidateSession_HappyPath() {
	userID := uuid.New()
	sessionID := uuid.New()
	user := models.User{ID: userID}
	session := models.Session{ID: sessionID}
	actor := models.UserActor{ID: userID, SessionID: sessionID}

	s.userRepo.On("GetByID", mock.Anything, userID).Return(user, nil)
	s.sessionRepo.On("GetByID", mock.Anything, sessionID).Return(session, nil)

	gotUser, gotSession, err := s.svc.ValidateSession(context.Background(), actor)

	require.NoError(s.T(), err)
	assert.Equal(s.T(), user, gotUser)
	assert.Equal(s.T(), session, gotSession)
}

func (s *AuthServiceSuite) TestValidateSession_UserNotFound() {
	userID := uuid.New()
	sessionID := uuid.New()
	actor := models.UserActor{ID: userID, SessionID: sessionID}

	s.userRepo.On("GetByID", mock.Anything, userID).Return(models.User{}, errx.ErrorUserNotFound)

	_, _, err := s.svc.ValidateSession(context.Background(), actor)

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, errx.ErrorUserInvalidSession)
}

func (s *AuthServiceSuite) TestValidateSession_UserRepoError() {
	userID := uuid.New()
	sessionID := uuid.New()
	repoErr := errors.New("db connection lost")
	actor := models.UserActor{ID: userID, SessionID: sessionID}

	s.userRepo.On("GetByID", mock.Anything, userID).Return(models.User{}, repoErr)

	_, _, err := s.svc.ValidateSession(context.Background(), actor)

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, repoErr)
}

func (s *AuthServiceSuite) TestValidateSession_SessionNotFound() {
	userID := uuid.New()
	sessionID := uuid.New()
	user := models.User{ID: userID}
	actor := models.UserActor{ID: userID, SessionID: sessionID}

	s.userRepo.On("GetByID", mock.Anything, userID).Return(user, nil)
	s.sessionRepo.On("GetByID", mock.Anything, sessionID).Return(models.Session{}, errx.ErrorSessionNotFound)

	_, _, err := s.svc.ValidateSession(context.Background(), actor)

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, errx.ErrorUserInvalidSession)
}

func (s *AuthServiceSuite) TestValidateSession_SessionRepoError() {
	userID := uuid.New()
	sessionID := uuid.New()
	user := models.User{ID: userID}
	repoErr := errors.New("db connection lost")
	actor := models.UserActor{ID: userID, SessionID: sessionID}

	s.userRepo.On("GetByID", mock.Anything, userID).Return(user, nil)
	s.sessionRepo.On("GetByID", mock.Anything, sessionID).Return(models.Session{}, repoErr)

	_, _, err := s.svc.ValidateSession(context.Background(), actor)

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, repoErr)
}
