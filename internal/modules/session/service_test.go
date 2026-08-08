package session

import (
	"context"
	"errors"
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/restkit/tokens"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type fakeTx struct{}

func (f *fakeTx) Transaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type SessionServiceSuite struct {
	suite.Suite

	auth          *mockAuth
	userRepo      *mockUserRepo
	emailRepo     *mockEmailRepo
	passwordRepo  *mockPasswordRepo
	sessionRepo   *mockSessionRepo
	userCache     *mockUserCache
	passwordCache *mockPasswordCache
	sessionsCache *mockSessionsCache
	passManager   *mockPasswordManager
	tokenManager  *mockTokenManager
	qrRepo        *mockQrRepo
	bus           *mockBus

	svc *Service
}

func (s *SessionServiceSuite) SetupTest() {
	s.auth = newMockAuth(s.T())
	s.userRepo = newMockUserRepo(s.T())
	s.emailRepo = newMockEmailRepo(s.T())
	s.passwordRepo = newMockPasswordRepo(s.T())
	s.sessionRepo = newMockSessionRepo(s.T())
	s.userCache = newMockUserCache(s.T())
	s.passwordCache = newMockPasswordCache(s.T())
	s.sessionsCache = newMockSessionsCache(s.T())
	s.passManager = newMockPasswordManager(s.T())
	s.tokenManager = newMockTokenManager(s.T())
	s.qrRepo = newMockQrRepo(s.T())
	s.bus = newMockBus(s.T())

	s.svc = New(ServiceDeps{
		Auth:          s.auth,
		UserRepo:      s.userRepo,
		EmailRepo:     s.emailRepo,
		PasswordRepo:  s.passwordRepo,
		SessionRepo:   s.sessionRepo,
		Tx:            &fakeTx{},
		PasswordCache: s.passwordCache,
		UserCache:     s.userCache,
		SessionsCache: s.sessionsCache,
		PassManager:   s.passManager,
		TokenManager:  s.tokenManager,
		QRStore:       s.qrRepo,
		Bus:           s.bus,
	})
}

func TestSessionService(t *testing.T) {
	suite.Run(t, new(SessionServiceSuite))
}

// ─── GetMySession ────────────────────────────────────────────────────────────

func (s *SessionServiceSuite) TestGetMySession_CacheHit() {
	userID := uuid.New()
	sessionID := uuid.New()
	actor := models.UserActor{ID: userID, SessionID: sessionID}
	session := models.Session{ID: sessionID, UserID: userID}

	s.sessionsCache.On("Get", mock.Anything, sessionID).Return(session, nil)

	got, err := s.svc.GetMySession(context.Background(), actor, sessionID)

	require.NoError(s.T(), err)
	assert.Equal(s.T(), session, got)
}

func (s *SessionServiceSuite) TestGetMySession_CacheHit_WrongUser() {
	userID := uuid.New()
	sessionID := uuid.New()
	actor := models.UserActor{ID: userID}
	session := models.Session{ID: sessionID, UserID: uuid.New()}

	s.sessionsCache.On("Get", mock.Anything, sessionID).Return(session, nil)

	_, err := s.svc.GetMySession(context.Background(), actor, sessionID)

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, errx.ErrorSessionNotFound)
}

func (s *SessionServiceSuite) TestGetMySession_CacheHit_Deleted() {
	userID := uuid.New()
	sessionID := uuid.New()
	actor := models.UserActor{ID: userID}
	deletedAt := time.Now()
	session := models.Session{ID: sessionID, UserID: userID, DeletedAt: &deletedAt}

	s.sessionsCache.On("Get", mock.Anything, sessionID).Return(session, nil)

	_, err := s.svc.GetMySession(context.Background(), actor, sessionID)

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, errx.ErrorSessionNotFound)
}

func (s *SessionServiceSuite) TestGetMySession_CacheMiss_RepoSuccess() {
	userID := uuid.New()
	sessionID := uuid.New()
	actor := models.UserActor{ID: userID}
	session := models.Session{ID: sessionID, UserID: userID}

	s.sessionsCache.On("Get", mock.Anything, sessionID).Return(models.Session{}, errors.New("miss"))
	s.sessionRepo.On("GetForUser", mock.Anything, userID, sessionID).Return(session, nil)
	s.sessionsCache.On("Set", mock.Anything, session).Return(nil).Maybe()

	got, err := s.svc.GetMySession(context.Background(), actor, sessionID)

	require.NoError(s.T(), err)
	assert.Equal(s.T(), session, got)
}

func (s *SessionServiceSuite) TestGetMySession_CacheMiss_RepoError() {
	userID := uuid.New()
	sessionID := uuid.New()
	repoErr := errors.New("db error")

	s.sessionsCache.On("Get", mock.Anything, sessionID).Return(models.Session{}, errors.New("miss"))
	s.sessionRepo.On("GetForUser", mock.Anything, userID, sessionID).Return(models.Session{}, repoErr)

	_, err := s.svc.GetMySession(context.Background(), models.UserActor{ID: userID}, sessionID)

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, repoErr)
}

// ─── Refresh ─────────────────────────────────────────────────────────────────

func (s *SessionServiceSuite) TestRefresh_ParseTokenError() {
	parseErr := errors.New("invalid token")
	s.tokenManager.On("ParseUserAuthRefresh", "bad_token").Return(tokens.AccountAuthClaims{}, parseErr)

	_, err := s.svc.Refresh(context.Background(), "bad_token")

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, errx.ErrorSessionExpired)
}

func (s *SessionServiceSuite) TestRefresh_GetTokenError() {
	sessionID := uuid.New()
	userID := uuid.New()
	claims := tokens.AccountAuthClaims{RegisteredClaims: jwtlib.RegisteredClaims{Subject: userID.String()}, SessionID: sessionID}
	repoErr := errors.New("db error")

	s.tokenManager.On("ParseUserAuthRefresh", "token").Return(claims, nil)
	s.sessionRepo.On("GetToken", mock.Anything, sessionID).Return("", repoErr)

	_, err := s.svc.Refresh(context.Background(), "token")

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, repoErr)
}

func (s *SessionServiceSuite) TestRefresh_HashError() {
	sessionID := uuid.New()
	userID := uuid.New()
	claims := tokens.AccountAuthClaims{RegisteredClaims: jwtlib.RegisteredClaims{Subject: userID.String()}, SessionID: sessionID}
	hashErr := errors.New("hash error")

	s.tokenManager.On("ParseUserAuthRefresh", "token").Return(claims, nil)
	s.sessionRepo.On("GetToken", mock.Anything, sessionID).Return("storedhash", nil)
	s.tokenManager.On("HashRefresh", "token").Return("", hashErr)

	_, err := s.svc.Refresh(context.Background(), "token")

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, hashErr)
}

func (s *SessionServiceSuite) TestRefresh_TokenMismatch() {
	sessionID := uuid.New()
	userID := uuid.New()
	claims := tokens.AccountAuthClaims{RegisteredClaims: jwtlib.RegisteredClaims{Subject: userID.String()}, SessionID: sessionID}

	s.tokenManager.On("ParseUserAuthRefresh", "token").Return(claims, nil)
	s.sessionRepo.On("GetToken", mock.Anything, sessionID).Return("storedhash", nil)
	s.tokenManager.On("HashRefresh", "token").Return("differenthash", nil)

	_, err := s.svc.Refresh(context.Background(), "token")

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, errx.ErrorSessionTokenMismatch)
}

func (s *SessionServiceSuite) TestRefresh_UserCacheHit() {
	sessionID := uuid.New()
	userID := uuid.New()
	claims := tokens.AccountAuthClaims{RegisteredClaims: jwtlib.RegisteredClaims{Subject: userID.String()}, SessionID: sessionID}
	user := models.User{ID: userID}
	session := models.Session{ID: sessionID}

	s.tokenManager.On("ParseUserAuthRefresh", "token").Return(claims, nil)
	s.sessionRepo.On("GetToken", mock.Anything, sessionID).Return("hash", nil)
	s.tokenManager.On("HashRefresh", "token").Return("hash", nil)
	s.userCache.On("Get", mock.Anything, userID).Return(user, nil)
	s.tokenManager.On("GenerateRefresh", user, sessionID).Return("newrefresh", nil)
	s.tokenManager.On("HashRefresh", "newrefresh").Return("newhash", nil)
	s.sessionRepo.On("UpdateToken", mock.Anything, sessionID, "newhash").Return(session, nil)
	s.tokenManager.On("GenerateAccess", user, sessionID).Return("access", nil)
	s.sessionsCache.On("Set", mock.Anything, session).Return(nil).Maybe()
	s.userCache.On("Set", mock.Anything, user).Return(nil).Maybe()

	pair, err := s.svc.Refresh(context.Background(), "token")

	require.NoError(s.T(), err)
	assert.Equal(s.T(), "newrefresh", pair.Refresh)
	assert.Equal(s.T(), "access", pair.Access)
}

func (s *SessionServiceSuite) TestRefresh_UserCacheMiss_RepoSuccess() {
	sessionID := uuid.New()
	userID := uuid.New()
	claims := tokens.AccountAuthClaims{RegisteredClaims: jwtlib.RegisteredClaims{Subject: userID.String()}, SessionID: sessionID}
	user := models.User{ID: userID}
	session := models.Session{ID: sessionID}

	s.tokenManager.On("ParseUserAuthRefresh", "token").Return(claims, nil)
	s.sessionRepo.On("GetToken", mock.Anything, sessionID).Return("hash", nil)
	s.tokenManager.On("HashRefresh", "token").Return("hash", nil)
	s.userCache.On("Get", mock.Anything, userID).Return(models.User{}, errors.New("miss"))
	s.userRepo.On("GetByID", mock.Anything, userID).Return(user, nil)
	s.tokenManager.On("GenerateRefresh", user, sessionID).Return("newrefresh", nil)
	s.tokenManager.On("HashRefresh", "newrefresh").Return("newhash", nil)
	s.sessionRepo.On("UpdateToken", mock.Anything, sessionID, "newhash").Return(session, nil)
	s.tokenManager.On("GenerateAccess", user, sessionID).Return("access", nil)
	s.sessionsCache.On("Set", mock.Anything, session).Return(nil).Maybe()
	s.userCache.On("Set", mock.Anything, user).Return(nil).Maybe()

	_, err := s.svc.Refresh(context.Background(), "token")

	require.NoError(s.T(), err)
}

func (s *SessionServiceSuite) TestRefresh_UserRepoError() {
	sessionID := uuid.New()
	userID := uuid.New()
	claims := tokens.AccountAuthClaims{RegisteredClaims: jwtlib.RegisteredClaims{Subject: userID.String()}, SessionID: sessionID}
	repoErr := errors.New("db error")

	s.tokenManager.On("ParseUserAuthRefresh", "token").Return(claims, nil)
	s.sessionRepo.On("GetToken", mock.Anything, sessionID).Return("hash", nil)
	s.tokenManager.On("HashRefresh", "token").Return("hash", nil)
	s.userCache.On("Get", mock.Anything, userID).Return(models.User{}, errors.New("miss"))
	s.userRepo.On("GetByID", mock.Anything, userID).Return(models.User{}, repoErr)

	_, err := s.svc.Refresh(context.Background(), "token")

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, repoErr)
}

func (s *SessionServiceSuite) TestRefresh_UpdateTokenError() {
	sessionID := uuid.New()
	userID := uuid.New()
	claims := tokens.AccountAuthClaims{RegisteredClaims: jwtlib.RegisteredClaims{Subject: userID.String()}, SessionID: sessionID}
	user := models.User{ID: userID}
	repoErr := errors.New("db error")

	s.tokenManager.On("ParseUserAuthRefresh", "token").Return(claims, nil)
	s.sessionRepo.On("GetToken", mock.Anything, sessionID).Return("hash", nil)
	s.tokenManager.On("HashRefresh", "token").Return("hash", nil)
	s.userCache.On("Get", mock.Anything, userID).Return(user, nil)
	s.tokenManager.On("GenerateRefresh", user, sessionID).Return("newrefresh", nil)
	s.tokenManager.On("HashRefresh", "newrefresh").Return("newhash", nil)
	s.sessionRepo.On("UpdateToken", mock.Anything, sessionID, "newhash").Return(models.Session{}, repoErr)

	_, err := s.svc.Refresh(context.Background(), "token")

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, repoErr)
}

// ─── Logout ──────────────────────────────────────────────────────────────────

func (s *SessionServiceSuite) TestLogout_RepoError() {
	actor := models.UserActor{ID: uuid.New(), SessionID: uuid.New()}
	repoErr := errors.New("db error")

	s.sessionRepo.On("Delete", mock.Anything, actor.SessionID).Return(repoErr)

	err := s.svc.Logout(context.Background(), actor)

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, repoErr)
}

func (s *SessionServiceSuite) TestLogout_HappyPath() {
	actor := models.UserActor{ID: uuid.New(), SessionID: uuid.New()}

	s.sessionRepo.On("Delete", mock.Anything, actor.SessionID).Return(nil)
	s.sessionsCache.On("Delete", mock.Anything, actor.SessionID).Return(nil).Maybe()

	err := s.svc.Logout(context.Background(), actor)

	require.NoError(s.T(), err)
}

// ─── DeleteMySession ─────────────────────────────────────────────────────────

func (s *SessionServiceSuite) TestDeleteMySession_ValidateSessionError() {
	actor := models.UserActor{ID: uuid.New(), SessionID: uuid.New()}
	sessionID := uuid.New()
	authErr := errors.New("invalid session")

	s.auth.On("ValidateSession", mock.Anything, actor).Return(models.User{}, models.Session{}, authErr)

	err := s.svc.DeleteMySession(context.Background(), actor, sessionID)

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, authErr)
}

func (s *SessionServiceSuite) TestDeleteMySession_RepoError() {
	actor := models.UserActor{ID: uuid.New(), SessionID: uuid.New()}
	sessionID := uuid.New()
	repoErr := errors.New("db error")

	s.auth.On("ValidateSession", mock.Anything, actor).Return(models.User{}, models.Session{}, nil)
	s.sessionRepo.On("DeleteOneForUser", mock.Anything, actor.ID, sessionID).Return(repoErr)

	err := s.svc.DeleteMySession(context.Background(), actor, sessionID)

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, repoErr)
}

func (s *SessionServiceSuite) TestDeleteMySession_HappyPath() {
	actor := models.UserActor{ID: uuid.New(), SessionID: uuid.New()}
	sessionID := uuid.New()

	s.auth.On("ValidateSession", mock.Anything, actor).Return(models.User{}, models.Session{}, nil)
	s.sessionRepo.On("DeleteOneForUser", mock.Anything, actor.ID, sessionID).Return(nil)
	s.sessionsCache.On("Delete", mock.Anything, sessionID).Return(nil).Maybe()

	err := s.svc.DeleteMySession(context.Background(), actor, sessionID)

	require.NoError(s.T(), err)
}

// ─── DeleteMySessions ────────────────────────────────────────────────────────

func (s *SessionServiceSuite) TestDeleteMySessions_ValidateSessionError() {
	actor := models.UserActor{ID: uuid.New(), SessionID: uuid.New()}
	authErr := errors.New("invalid session")

	s.auth.On("ValidateSession", mock.Anything, actor).Return(models.User{}, models.Session{}, authErr)

	err := s.svc.DeleteMySessions(context.Background(), actor)

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, authErr)
}

func (s *SessionServiceSuite) TestDeleteMySessions_RepoError() {
	actor := models.UserActor{ID: uuid.New(), SessionID: uuid.New()}
	repoErr := errors.New("db error")

	s.auth.On("ValidateSession", mock.Anything, actor).Return(models.User{}, models.Session{}, nil)
	s.sessionRepo.On("DeleteManyForUser", mock.Anything, actor.ID).Return([]uuid.UUID(nil), repoErr)

	err := s.svc.DeleteMySessions(context.Background(), actor)

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, repoErr)
}

func (s *SessionServiceSuite) TestDeleteMySessions_HappyPath() {
	actor := models.UserActor{ID: uuid.New(), SessionID: uuid.New()}
	id1, id2 := uuid.New(), uuid.New()

	s.auth.On("ValidateSession", mock.Anything, actor).Return(models.User{}, models.Session{}, nil)
	s.sessionRepo.On("DeleteManyForUser", mock.Anything, actor.ID).Return([]uuid.UUID{id1, id2}, nil)
	s.sessionsCache.On("Delete", mock.Anything, id1).Return(nil).Maybe()
	s.sessionsCache.On("Delete", mock.Anything, id2).Return(nil).Maybe()

	err := s.svc.DeleteMySessions(context.Background(), actor)

	require.NoError(s.T(), err)
}

// ─── LoginByEmail ────────────────────────────────────────────────────────────

func (s *SessionServiceSuite) TestLoginByEmail_EmailRepoError() {
	repoErr := errors.New("email not found")
	s.emailRepo.On("GetByEmail", mock.Anything, "user@example.com").Return(models.UserEmail{}, repoErr)

	_, err := s.svc.LoginByEmail(context.Background(), "user@example.com", "Password1!")

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, repoErr)
}

func (s *SessionServiceSuite) TestLoginByEmail_UserRepoError() {
	userID := uuid.New()
	emailRecord := models.UserEmail{UserID: userID}
	repoErr := errors.New("db error")

	s.emailRepo.On("GetByEmail", mock.Anything, "user@example.com").Return(emailRecord, nil)
	s.userRepo.On("GetByID", mock.Anything, userID).Return(models.User{}, repoErr)

	_, err := s.svc.LoginByEmail(context.Background(), "user@example.com", "Password1!")

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, repoErr)
}

func (s *SessionServiceSuite) TestLoginByEmail_WrongPassword() {
	userID := uuid.New()
	emailRecord := models.UserEmail{UserID: userID}
	user := models.User{ID: userID}
	pwd := models.UserPassword{UserID: userID, Hash: "hash"}
	checkErr := errors.New("password mismatch")

	s.emailRepo.On("GetByEmail", mock.Anything, "user@example.com").Return(emailRecord, nil)
	s.userRepo.On("GetByID", mock.Anything, userID).Return(user, nil)
	s.passwordCache.On("Get", mock.Anything, userID).Return(pwd, nil)
	s.passManager.On("CheckMatch", "wrongpass", pwd.Hash).Return(checkErr)

	_, err := s.svc.LoginByEmail(context.Background(), "user@example.com", "wrongpass")

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, checkErr)
}

func (s *SessionServiceSuite) TestLoginByEmail_HappyPath() {
	userID := uuid.New()
	sessionID := uuid.New()
	emailRecord := models.UserEmail{UserID: userID}
	user := models.User{ID: userID}
	pwd := models.UserPassword{UserID: userID, Hash: "hash"}
	session := models.Session{ID: sessionID, UserID: userID}

	s.emailRepo.On("GetByEmail", mock.Anything, "user@example.com").Return(emailRecord, nil)
	s.userRepo.On("GetByID", mock.Anything, userID).Return(user, nil)
	s.passwordCache.On("Get", mock.Anything, userID).Return(pwd, nil)
	s.passManager.On("CheckMatch", "Password1!", pwd.Hash).Return(nil)
	s.tokenManager.On("GenerateRefresh", user, mock.Anything).Return("refresh", nil)
	s.tokenManager.On("HashRefresh", "refresh").Return("hash", nil)
	s.sessionRepo.On("Create", mock.Anything, mock.Anything, userID, "hash").Return(session, nil)
	s.tokenManager.On("GenerateAccess", user, session.ID).Return("access", nil)
	s.userCache.On("Set", mock.Anything, user).Return(nil).Maybe()
	s.sessionsCache.On("Set", mock.Anything, session).Return(nil).Maybe()

	pair, err := s.svc.LoginByEmail(context.Background(), "user@example.com", "Password1!")

	require.NoError(s.T(), err)
	assert.Equal(s.T(), "refresh", pair.Refresh)
	assert.Equal(s.T(), "access", pair.Access)
}

// ─── LoginByGoogle ───────────────────────────────────────────────────────────

func (s *SessionServiceSuite) TestLoginByGoogle_EmailRepoError() {
	repoErr := errors.New("email not found")
	s.emailRepo.On("GetByEmail", mock.Anything, "user@gmail.com").Return(models.UserEmail{}, repoErr)

	_, err := s.svc.LoginByGoogle(context.Background(), "user@gmail.com")

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, repoErr)
}

func (s *SessionServiceSuite) TestLoginByGoogle_UserRepoError() {
	userID := uuid.New()
	emailRecord := models.UserEmail{UserID: userID}
	repoErr := errors.New("db error")

	s.emailRepo.On("GetByEmail", mock.Anything, "user@gmail.com").Return(emailRecord, nil)
	s.userRepo.On("GetByID", mock.Anything, userID).Return(models.User{}, repoErr)

	_, err := s.svc.LoginByGoogle(context.Background(), "user@gmail.com")

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, repoErr)
}

func (s *SessionServiceSuite) TestLoginByGoogle_HappyPath() {
	userID := uuid.New()
	emailRecord := models.UserEmail{UserID: userID}
	user := models.User{ID: userID}
	session := models.Session{ID: uuid.New(), UserID: userID}

	s.emailRepo.On("GetByEmail", mock.Anything, "user@gmail.com").Return(emailRecord, nil)
	s.userRepo.On("GetByID", mock.Anything, userID).Return(user, nil)
	s.tokenManager.On("GenerateRefresh", user, mock.Anything).Return("refresh", nil)
	s.tokenManager.On("HashRefresh", "refresh").Return("hash", nil)
	s.sessionRepo.On("Create", mock.Anything, mock.Anything, userID, "hash").Return(session, nil)
	s.tokenManager.On("GenerateAccess", user, session.ID).Return("access", nil)
	s.userCache.On("Set", mock.Anything, user).Return(nil).Maybe()
	s.sessionsCache.On("Set", mock.Anything, session).Return(nil).Maybe()

	pair, err := s.svc.LoginByGoogle(context.Background(), "user@gmail.com")

	require.NoError(s.T(), err)
	assert.Equal(s.T(), "refresh", pair.Refresh)
}

// ─── checkPassword (через LoginByEmail) ──────────────────────────────────────

func (s *SessionServiceSuite) TestCheckPassword_CacheMiss_RepoSuccess() {
	userID := uuid.New()
	emailRecord := models.UserEmail{UserID: userID}
	user := models.User{ID: userID}
	pwd := models.UserPassword{UserID: userID, Hash: "hash"}
	session := models.Session{ID: uuid.New()}

	s.emailRepo.On("GetByEmail", mock.Anything, mock.Anything).Return(emailRecord, nil)
	s.userRepo.On("GetByID", mock.Anything, userID).Return(user, nil)
	s.passwordCache.On("Get", mock.Anything, userID).Return(models.UserPassword{}, errors.New("miss"))
	s.passwordRepo.On("GetByID", mock.Anything, userID).Return(pwd, nil)
	s.passwordCache.On("Set", mock.Anything, pwd).Return(nil).Maybe()
	s.passManager.On("CheckMatch", "Password1!", pwd.Hash).Return(nil)
	s.tokenManager.On("GenerateRefresh", user, mock.Anything).Return("refresh", nil)
	s.tokenManager.On("HashRefresh", "refresh").Return("hash", nil)
	s.sessionRepo.On("Create", mock.Anything, mock.Anything, userID, "hash").Return(session, nil)
	s.tokenManager.On("GenerateAccess", user, session.ID).Return("access", nil)
	s.userCache.On("Set", mock.Anything, user).Return(nil).Maybe()
	s.sessionsCache.On("Set", mock.Anything, session).Return(nil).Maybe()

	_, err := s.svc.LoginByEmail(context.Background(), "user@example.com", "Password1!")

	require.NoError(s.T(), err)
}

func (s *SessionServiceSuite) TestCheckPassword_CacheMiss_RepoError() {
	userID := uuid.New()
	emailRecord := models.UserEmail{UserID: userID}
	user := models.User{ID: userID}
	repoErr := errors.New("db error")

	s.emailRepo.On("GetByEmail", mock.Anything, mock.Anything).Return(emailRecord, nil)
	s.userRepo.On("GetByID", mock.Anything, userID).Return(user, nil)
	s.passwordCache.On("Get", mock.Anything, userID).Return(models.UserPassword{}, errors.New("miss"))
	s.passwordRepo.On("GetByID", mock.Anything, userID).Return(models.UserPassword{}, repoErr)

	_, err := s.svc.LoginByEmail(context.Background(), "user@example.com", "Password1!")

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, repoErr)
}

// ─── CreateQRToken ───────────────────────────────────────────────────────────

func (s *SessionServiceSuite) TestCreateQRToken_RepoError() {
	repoErr := errors.New("redis error")
	s.qrRepo.On("Set", mock.Anything, mock.Anything, "pending", QRTokenTTL).Return(repoErr)

	_, err := s.svc.CreateQRToken(context.Background())

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, repoErr)
}

func (s *SessionServiceSuite) TestCreateQRToken_HappyPath() {
	s.qrRepo.On("Set", mock.Anything, mock.Anything, "pending", QRTokenTTL).Return(nil)

	token, err := s.svc.CreateQRToken(context.Background())

	require.NoError(s.T(), err)
	assert.NotEmpty(s.T(), token)
}

// ─── ConfirmQRToken ──────────────────────────────────────────────────────────

func (s *SessionServiceSuite) TestConfirmQRToken_NotFound() {
	actor := models.UserActor{ID: uuid.New()}

	s.qrRepo.On("Get", mock.Anything, "qr_token").Return("", errx.ErrorQRTokenNotFound)

	_, err := s.svc.ConfirmQRToken(context.Background(), actor, "qr_token")

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, errx.ErrorQRTokenNotFound)
}

func (s *SessionServiceSuite) TestConfirmQRToken_AlreadyConfirmed() {
	actor := models.UserActor{ID: uuid.New()}

	s.qrRepo.On("Get", mock.Anything, "qr_token").Return("confirmed", nil)

	_, err := s.svc.ConfirmQRToken(context.Background(), actor, "qr_token")

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, errx.ErrorQRTokenAlreadyConfirmed)
}

func (s *SessionServiceSuite) TestConfirmQRToken_UserRepoError() {
	actor := models.UserActor{ID: uuid.New()}
	repoErr := errors.New("db error")

	s.qrRepo.On("Get", mock.Anything, "qr_token").Return("pending", nil)
	s.userRepo.On("GetByID", mock.Anything, actor.ID).Return(models.User{}, repoErr)

	_, err := s.svc.ConfirmQRToken(context.Background(), actor, "qr_token")

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, repoErr)
}

func (s *SessionServiceSuite) TestConfirmQRToken_HappyPath() {
	actor := models.UserActor{ID: uuid.New()}
	user := models.User{ID: actor.ID}
	session := models.Session{ID: uuid.New()}

	s.qrRepo.On("Get", mock.Anything, "qr_token").Return("pending", nil)
	s.userRepo.On("GetByID", mock.Anything, actor.ID).Return(user, nil)
	s.tokenManager.On("GenerateRefresh", user, mock.Anything).Return("refresh", nil)
	s.tokenManager.On("HashRefresh", "refresh").Return("hash", nil)
	s.sessionRepo.On("Create", mock.Anything, mock.Anything, actor.ID, "hash").Return(session, nil)
	s.tokenManager.On("GenerateAccess", user, session.ID).Return("access", nil)
	s.qrRepo.On("Set", mock.Anything, "qr_token", "confirmed", qrConfirmedTTL).Return(nil)
	s.userCache.On("Set", mock.Anything, user).Return(nil).Maybe()
	s.sessionsCache.On("Set", mock.Anything, session).Return(nil).Maybe()

	pair, err := s.svc.ConfirmQRToken(context.Background(), actor, "qr_token")

	require.NoError(s.T(), err)
	assert.Equal(s.T(), "refresh", pair.Refresh)
}

// ─── PublishQRToken ──────────────────────────────────────────────────────────

func (s *SessionServiceSuite) TestPublishQRToken_DelegatesToBus() {
	payload := []byte(`{"token":"abc"}`)
	s.bus.On("PublishQRToken", mock.Anything, "key", payload).Return(nil)

	err := s.svc.PublishQRToken(context.Background(), "key", payload)

	require.NoError(s.T(), err)
}
