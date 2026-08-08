package user

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

type fakeTx struct{}

func (f *fakeTx) Transaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type UserServiceSuite struct {
	suite.Suite

	auth          *mockAuth
	userRepo      *mockUserRepo
	emailRepo     *mockEmailRepo
	passwordRepo  *mockPasswordRepo
	sessionRepo   *mockSessionRepo
	userCache     *mockUserCache
	emailCache    *mockEmailCache
	passwordCache *mockPasswordCache
	sessionsCache *mockSessionsCache
	passManager   *mockPasswordManager
	messenger     *mockMessenger

	svc *Service
}

func (s *UserServiceSuite) SetupTest() {
	s.auth = newMockAuth(s.T())
	s.userRepo = newMockUserRepo(s.T())
	s.emailRepo = newMockEmailRepo(s.T())
	s.passwordRepo = newMockPasswordRepo(s.T())
	s.sessionRepo = newMockSessionRepo(s.T())
	s.userCache = newMockUserCache(s.T())
	s.emailCache = newMockEmailCache(s.T())
	s.passwordCache = newMockPasswordCache(s.T())
	s.sessionsCache = newMockSessionsCache(s.T())
	s.passManager = newMockPasswordManager(s.T())
	s.messenger = newMockMessenger(s.T())

	s.svc = New(ServiceDeps{
		Auth:          s.auth,
		UserRepo:      s.userRepo,
		EmailRepo:     s.emailRepo,
		PasswordRepo:  s.passwordRepo,
		SessionRepo:   s.sessionRepo,
		Tx:            &fakeTx{},
		UserCache:     s.userCache,
		EmailCache:    s.emailCache,
		PasswordCache: s.passwordCache,
		SessionsCache: s.sessionsCache,
		PassManager:   s.passManager,
		Messenger:     s.messenger,
	})
}

func TestUserService(t *testing.T) {
	suite.Run(t, new(UserServiceSuite))
}

// ─── Registration ───────────────────────────────────────────────────────────

func (s *UserServiceSuite) TestRegistration_HappyPath() {
	ctx := context.Background()
	userID := uuid.New()

	params := RegistrationParams{
		Email:    "user@example.com",
		Password: "Password1!",
		Role:     "user",
	}

	user := models.User{ID: userID}
	email := models.UserEmail{UserID: userID, Email: params.Email}
	password := models.UserPassword{UserID: userID, Hash: "hash"}

	s.passManager.On("GenerateHash", params.Password).Return("hash", nil)
	s.userRepo.On("Create", mock.Anything, params).Return(user, nil)
	s.emailRepo.On("Create", mock.Anything, models.UserEmail{UserID: userID, Email: params.Email}).Return(email, nil)
	s.passwordRepo.On("Create", mock.Anything, models.UserPassword{UserID: userID, Hash: "hash"}).Return(password, nil)
	s.messenger.On("WriteUserCreated", mock.Anything, user, email).Return(nil)
	s.userCache.On("Set", mock.Anything, user).Return(nil).Maybe()
	s.emailCache.On("Set", mock.Anything, email).Return(nil).Maybe()
	s.passwordCache.On("Set", mock.Anything, password).Return(nil).Maybe()

	got, err := s.svc.Registration(ctx, params)

	require.NoError(s.T(), err)
	assert.Equal(s.T(), user, got)
}

func (s *UserServiceSuite) TestRegistration_InvalidRole() {
	_, err := s.svc.Registration(context.Background(), RegistrationParams{
		Role:     "superadmin",
		Password: "Password1!",
	})

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, errx.ErrorRoleNotSupported)
}

func (s *UserServiceSuite) TestRegistration_InvalidPassword_TooShort() {
	_, err := s.svc.Registration(context.Background(), RegistrationParams{
		Role:     "user",
		Password: "Ab1!",
	})

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, errx.ErrorPasswordIsNotAllowed)
}

func (s *UserServiceSuite) TestRegistration_InvalidPassword_MissingSpecial() {
	_, err := s.svc.Registration(context.Background(), RegistrationParams{
		Role:     "user",
		Password: "Password1",
	})

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, errx.ErrorPasswordIsNotAllowed)
}

func (s *UserServiceSuite) TestRegistration_GenerateHashError() {
	hashErr := errors.New("bcrypt error")
	s.passManager.On("GenerateHash", mock.Anything).Return("", hashErr)

	_, err := s.svc.Registration(context.Background(), RegistrationParams{
		Role:     "user",
		Password: "Password1!",
	})

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, hashErr)
}

func (s *UserServiceSuite) TestRegistration_UserRepoError() {
	repoErr := errors.New("db error")
	s.passManager.On("GenerateHash", mock.Anything).Return("hash", nil)
	s.userRepo.On("Create", mock.Anything, mock.Anything).Return(models.User{}, repoErr)

	_, err := s.svc.Registration(context.Background(), RegistrationParams{
		Role:     "user",
		Password: "Password1!",
	})

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, repoErr)
}

func (s *UserServiceSuite) TestRegistration_EmailRepoError() {
	repoErr := errors.New("db error")
	user := models.User{ID: uuid.New()}
	s.passManager.On("GenerateHash", mock.Anything).Return("hash", nil)
	s.userRepo.On("Create", mock.Anything, mock.Anything).Return(user, nil)
	s.emailRepo.On("Create", mock.Anything, mock.Anything).Return(models.UserEmail{}, repoErr)

	_, err := s.svc.Registration(context.Background(), RegistrationParams{
		Role:     "user",
		Password: "Password1!",
	})

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, repoErr)
}

func (s *UserServiceSuite) TestRegistration_PasswordRepoError() {
	repoErr := errors.New("db error")
	user := models.User{ID: uuid.New()}
	email := models.UserEmail{UserID: user.ID}
	s.passManager.On("GenerateHash", mock.Anything).Return("hash", nil)
	s.userRepo.On("Create", mock.Anything, mock.Anything).Return(user, nil)
	s.emailRepo.On("Create", mock.Anything, mock.Anything).Return(email, nil)
	s.passwordRepo.On("Create", mock.Anything, mock.Anything).Return(models.UserPassword{}, repoErr)

	_, err := s.svc.Registration(context.Background(), RegistrationParams{
		Role:     "user",
		Password: "Password1!",
	})

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, repoErr)
}

func (s *UserServiceSuite) TestRegistration_MessengerError() {
	msgErr := errors.New("kafka error")
	user := models.User{ID: uuid.New()}
	email := models.UserEmail{UserID: user.ID}
	password := models.UserPassword{UserID: user.ID}
	s.passManager.On("GenerateHash", mock.Anything).Return("hash", nil)
	s.userRepo.On("Create", mock.Anything, mock.Anything).Return(user, nil)
	s.emailRepo.On("Create", mock.Anything, mock.Anything).Return(email, nil)
	s.passwordRepo.On("Create", mock.Anything, mock.Anything).Return(password, nil)
	s.messenger.On("WriteUserCreated", mock.Anything, user, email).Return(msgErr)

	_, err := s.svc.Registration(context.Background(), RegistrationParams{
		Role:     "user",
		Password: "Password1!",
	})

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, msgErr)
}

// ─── GetMyUserByID ────────────────────────────────────────────────────────

func (s *UserServiceSuite) TestGetMyUserByID_CacheHit() {
	userID := uuid.New()
	user := models.User{ID: userID}

	s.userCache.On("Get", mock.Anything, userID).Return(user, nil)

	got, err := s.svc.GetMyUserByID(context.Background(), models.UserActor{ID: userID})

	require.NoError(s.T(), err)
	assert.Equal(s.T(), user, got)
}

func (s *UserServiceSuite) TestGetMyUserByID_CacheMiss_RepoSuccess() {
	userID := uuid.New()
	user := models.User{ID: userID}

	s.userCache.On("Get", mock.Anything, userID).Return(models.User{}, errors.New("miss"))
	s.userRepo.On("GetByID", mock.Anything, userID).Return(user, nil)
	s.userCache.On("Set", mock.Anything, user).Return(nil).Maybe()

	got, err := s.svc.GetMyUserByID(context.Background(), models.UserActor{ID: userID})

	require.NoError(s.T(), err)
	assert.Equal(s.T(), user, got)
}

func (s *UserServiceSuite) TestGetMyUserByID_CacheMiss_RepoError() {
	userID := uuid.New()
	repoErr := errors.New("db error")

	s.userCache.On("Get", mock.Anything, userID).Return(models.User{}, errors.New("miss"))
	s.userRepo.On("GetByID", mock.Anything, userID).Return(models.User{}, repoErr)

	_, err := s.svc.GetMyUserByID(context.Background(), models.UserActor{ID: userID})

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, repoErr)
}

// ─── GetMyEmailByID ──────────────────────────────────────────────────────────

func (s *UserServiceSuite) TestGetMyEmailByID_CacheHit() {
	userID := uuid.New()
	email := models.UserEmail{UserID: userID}

	s.emailCache.On("GetByID", mock.Anything, userID).Return(email, nil)

	got, err := s.svc.GetMyEmailByID(context.Background(), models.UserActor{ID: userID})

	require.NoError(s.T(), err)
	assert.Equal(s.T(), email, got)
}

func (s *UserServiceSuite) TestGetMyEmailByID_CacheMiss_RepoSuccess() {
	userID := uuid.New()
	email := models.UserEmail{UserID: userID}

	s.emailCache.On("GetByID", mock.Anything, userID).Return(models.UserEmail{}, errors.New("miss"))
	s.emailRepo.On("GetByID", mock.Anything, userID).Return(email, nil)
	s.emailCache.On("Set", mock.Anything, email).Return(nil).Maybe()

	got, err := s.svc.GetMyEmailByID(context.Background(), models.UserActor{ID: userID})

	require.NoError(s.T(), err)
	assert.Equal(s.T(), email, got)
}

func (s *UserServiceSuite) TestGetMyEmailByID_CacheMiss_RepoError() {
	userID := uuid.New()
	repoErr := errors.New("db error")

	s.emailCache.On("GetByID", mock.Anything, userID).Return(models.UserEmail{}, errors.New("miss"))
	s.emailRepo.On("GetByID", mock.Anything, userID).Return(models.UserEmail{}, repoErr)

	_, err := s.svc.GetMyEmailByID(context.Background(), models.UserActor{ID: userID})

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, repoErr)
}

// ─── UpdatePassword ──────────────────────────────────────────────────────────

func (s *UserServiceSuite) TestUpdatePassword_ValidateSessionError() {
	authErr := errors.New("session invalid")
	actor := models.UserActor{ID: uuid.New(), SessionID: uuid.New()}

	s.auth.On("ValidateSession", mock.Anything, actor).Return(models.User{}, models.Session{}, authErr)

	err := s.svc.UpdatePassword(context.Background(), actor, "old", "new")

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, authErr)
}

func (s *UserServiceSuite) TestUpdatePassword_PasswordFromCache() {
	actor := models.UserActor{ID: uuid.New(), SessionID: uuid.New()}
	pwd := models.UserPassword{UserID: actor.ID, Hash: "oldhash"}
	updated := models.UserPassword{UserID: actor.ID, Hash: "newhash"}

	s.auth.On("ValidateSession", mock.Anything, actor).Return(models.User{}, models.Session{}, nil)
	s.passwordCache.On("Get", mock.Anything, actor.ID).Return(pwd, nil)
	s.passManager.On("CheckMatch", "OldPass1!", pwd.Hash).Return(nil)
	s.passManager.On("GenerateHash", "NewPass1!").Return("newhash", nil)
	s.passwordRepo.On("UpdatePassword", mock.Anything, actor.ID, "newhash").Return(updated, nil)
	s.passwordCache.On("Set", mock.Anything, updated).Return(nil).Maybe()

	err := s.svc.UpdatePassword(context.Background(), actor, "OldPass1!", "NewPass1!")

	require.NoError(s.T(), err)
}

func (s *UserServiceSuite) TestUpdatePassword_PasswordFromRepo() {
	actor := models.UserActor{ID: uuid.New(), SessionID: uuid.New()}
	pwd := models.UserPassword{UserID: actor.ID, Hash: "oldhash"}
	updated := models.UserPassword{UserID: actor.ID, Hash: "newhash"}

	s.auth.On("ValidateSession", mock.Anything, actor).Return(models.User{}, models.Session{}, nil)
	s.passwordCache.On("Get", mock.Anything, actor.ID).Return(models.UserPassword{}, errors.New("miss"))
	s.passwordRepo.On("GetByID", mock.Anything, actor.ID).Return(pwd, nil)
	s.passManager.On("CheckMatch", "OldPass1!", pwd.Hash).Return(nil)
	s.passManager.On("GenerateHash", "NewPass1!").Return("newhash", nil)
	s.passwordRepo.On("UpdatePassword", mock.Anything, actor.ID, "newhash").Return(updated, nil)
	s.passwordCache.On("Set", mock.Anything, updated).Return(nil).Maybe()

	err := s.svc.UpdatePassword(context.Background(), actor, "OldPass1!", "NewPass1!")

	require.NoError(s.T(), err)
}

func (s *UserServiceSuite) TestUpdatePassword_RepoGetError() {
	actor := models.UserActor{ID: uuid.New(), SessionID: uuid.New()}
	repoErr := errors.New("db error")

	s.auth.On("ValidateSession", mock.Anything, actor).Return(models.User{}, models.Session{}, nil)
	s.passwordCache.On("Get", mock.Anything, actor.ID).Return(models.UserPassword{}, errors.New("miss"))
	s.passwordRepo.On("GetByID", mock.Anything, actor.ID).Return(models.UserPassword{}, repoErr)

	err := s.svc.UpdatePassword(context.Background(), actor, "OldPass1!", "NewPass1!")

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, repoErr)
}

func (s *UserServiceSuite) TestUpdatePassword_WrongOldPassword() {
	actor := models.UserActor{ID: uuid.New(), SessionID: uuid.New()}
	pwd := models.UserPassword{UserID: actor.ID, Hash: "oldhash"}
	checkErr := errors.New("password mismatch")

	s.auth.On("ValidateSession", mock.Anything, actor).Return(models.User{}, models.Session{}, nil)
	s.passwordCache.On("Get", mock.Anything, actor.ID).Return(pwd, nil)
	s.passManager.On("CheckMatch", "WrongPass1!", pwd.Hash).Return(checkErr)

	err := s.svc.UpdatePassword(context.Background(), actor, "WrongPass1!", "NewPass1!")

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, checkErr)
}

func (s *UserServiceSuite) TestUpdatePassword_InvalidNewPassword() {
	actor := models.UserActor{ID: uuid.New(), SessionID: uuid.New()}
	pwd := models.UserPassword{UserID: actor.ID, Hash: "oldhash"}

	s.auth.On("ValidateSession", mock.Anything, actor).Return(models.User{}, models.Session{}, nil)
	s.passwordCache.On("Get", mock.Anything, actor.ID).Return(pwd, nil)
	s.passManager.On("CheckMatch", "OldPass1!", pwd.Hash).Return(nil)

	err := s.svc.UpdatePassword(context.Background(), actor, "OldPass1!", "short")

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, errx.ErrorPasswordIsNotAllowed)
}

func (s *UserServiceSuite) TestUpdatePassword_GenerateHashError() {
	actor := models.UserActor{ID: uuid.New(), SessionID: uuid.New()}
	pwd := models.UserPassword{UserID: actor.ID, Hash: "oldhash"}
	hashErr := errors.New("bcrypt error")

	s.auth.On("ValidateSession", mock.Anything, actor).Return(models.User{}, models.Session{}, nil)
	s.passwordCache.On("Get", mock.Anything, actor.ID).Return(pwd, nil)
	s.passManager.On("CheckMatch", "OldPass1!", pwd.Hash).Return(nil)
	s.passManager.On("GenerateHash", "NewPass1!").Return("", hashErr)

	err := s.svc.UpdatePassword(context.Background(), actor, "OldPass1!", "NewPass1!")

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, hashErr)
}

func (s *UserServiceSuite) TestUpdatePassword_UpdateRepoError() {
	actor := models.UserActor{ID: uuid.New(), SessionID: uuid.New()}
	pwd := models.UserPassword{UserID: actor.ID, Hash: "oldhash"}
	repoErr := errors.New("db error")

	s.auth.On("ValidateSession", mock.Anything, actor).Return(models.User{}, models.Session{}, nil)
	s.passwordCache.On("Get", mock.Anything, actor.ID).Return(pwd, nil)
	s.passManager.On("CheckMatch", "OldPass1!", pwd.Hash).Return(nil)
	s.passManager.On("GenerateHash", "NewPass1!").Return("newhash", nil)
	s.passwordRepo.On("UpdatePassword", mock.Anything, actor.ID, "newhash").Return(models.UserPassword{}, repoErr)

	err := s.svc.UpdatePassword(context.Background(), actor, "OldPass1!", "NewPass1!")

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, repoErr)
}

// ─── DeleteMyUser ─────────────────────────────────────────────────────────

func (s *UserServiceSuite) TestDeleteMyUser_ValidateSessionError() {
	actor := models.UserActor{ID: uuid.New(), SessionID: uuid.New()}
	authErr := errors.New("session invalid")

	s.auth.On("ValidateSession", mock.Anything, actor).Return(models.User{}, models.Session{}, authErr)

	err := s.svc.DeleteMyUser(context.Background(), actor)

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, authErr)
}

func (s *UserServiceSuite) TestDeleteMyUser_UserRepoError() {
	actor := models.UserActor{ID: uuid.New(), SessionID: uuid.New()}
	repoErr := errors.New("db error")

	s.auth.On("ValidateSession", mock.Anything, actor).Return(models.User{}, models.Session{}, nil)
	s.userRepo.On("Delete", mock.Anything, actor.ID).Return(models.User{}, repoErr)

	err := s.svc.DeleteMyUser(context.Background(), actor)

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, repoErr)
}

func (s *UserServiceSuite) TestDeleteMyUser_EmailRepoError() {
	actor := models.UserActor{ID: uuid.New(), SessionID: uuid.New()}
	user := models.User{ID: actor.ID}
	repoErr := errors.New("db error")

	s.auth.On("ValidateSession", mock.Anything, actor).Return(models.User{}, models.Session{}, nil)
	s.userRepo.On("Delete", mock.Anything, actor.ID).Return(user, nil)
	s.emailRepo.On("GetByID", mock.Anything, actor.ID, mock.Anything).Return(models.UserEmail{}, repoErr)

	err := s.svc.DeleteMyUser(context.Background(), actor)

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, repoErr)
}

func (s *UserServiceSuite) TestDeleteMyUser_SessionRepoError() {
	actor := models.UserActor{ID: uuid.New(), SessionID: uuid.New()}
	user := models.User{ID: actor.ID}
	email := models.UserEmail{UserID: actor.ID}
	repoErr := errors.New("db error")

	s.auth.On("ValidateSession", mock.Anything, actor).Return(models.User{}, models.Session{}, nil)
	s.userRepo.On("Delete", mock.Anything, actor.ID).Return(user, nil)
	s.emailRepo.On("GetByID", mock.Anything, actor.ID, mock.Anything).Return(email, nil)
	s.sessionRepo.On("DeleteManyForUser", mock.Anything, actor.ID).Return([]uuid.UUID(nil), repoErr)

	err := s.svc.DeleteMyUser(context.Background(), actor)

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, repoErr)
}

func (s *UserServiceSuite) TestDeleteMyUser_MessengerError() {
	actor := models.UserActor{ID: uuid.New(), SessionID: uuid.New()}
	user := models.User{ID: actor.ID}
	email := models.UserEmail{UserID: actor.ID}
	msgErr := errors.New("kafka error")

	s.auth.On("ValidateSession", mock.Anything, actor).Return(models.User{}, models.Session{}, nil)
	s.userRepo.On("Delete", mock.Anything, actor.ID).Return(user, nil)
	s.emailRepo.On("GetByID", mock.Anything, actor.ID, mock.Anything).Return(email, nil)
	s.sessionRepo.On("DeleteManyForUser", mock.Anything, actor.ID).Return([]uuid.UUID{}, nil)
	s.messenger.On("WriteUserDeleted", mock.Anything, user, email).Return(msgErr)

	err := s.svc.DeleteMyUser(context.Background(), actor)

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, msgErr)
}

func (s *UserServiceSuite) TestDeleteMyUser_HappyPath() {
	actor := models.UserActor{ID: uuid.New(), SessionID: uuid.New()}
	user := models.User{ID: actor.ID}
	email := models.UserEmail{UserID: actor.ID}
	sessionID := uuid.New()

	s.auth.On("ValidateSession", mock.Anything, actor).Return(models.User{}, models.Session{}, nil)
	s.userRepo.On("Delete", mock.Anything, actor.ID).Return(user, nil)
	s.emailRepo.On("GetByID", mock.Anything, actor.ID, mock.Anything).Return(email, nil)
	s.sessionRepo.On("DeleteManyForUser", mock.Anything, actor.ID).Return([]uuid.UUID{sessionID}, nil)
	s.messenger.On("WriteUserDeleted", mock.Anything, user, email).Return(nil)
	s.userCache.On("Delete", mock.Anything, actor.ID).Return(nil).Maybe()
	s.emailCache.On("DeleteByID", mock.Anything, actor.ID).Return(nil).Maybe()
	s.passwordCache.On("Delete", mock.Anything, actor.ID).Return(nil).Maybe()
	s.sessionsCache.On("Delete", mock.Anything, sessionID).Return(nil).Maybe()

	err := s.svc.DeleteMyUser(context.Background(), actor)

	require.NoError(s.T(), err)
}
