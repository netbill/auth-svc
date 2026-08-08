package account

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

type AccountServiceSuite struct {
	suite.Suite

	auth          *mockAuth
	accountRepo   *mockAccountRepo
	emailRepo     *mockEmailRepo
	passwordRepo  *mockPasswordRepo
	sessionRepo   *mockSessionRepo
	accountCache  *mockAccountCache
	emailCache    *mockEmailCache
	passwordCache *mockPasswordCache
	sessionsCache *mockSessionsCache
	passManager   *mockPasswordManager
	messenger     *mockMessenger

	svc *Service
}

func (s *AccountServiceSuite) SetupTest() {
	s.auth = newMockAuth(s.T())
	s.accountRepo = newMockAccountRepo(s.T())
	s.emailRepo = newMockEmailRepo(s.T())
	s.passwordRepo = newMockPasswordRepo(s.T())
	s.sessionRepo = newMockSessionRepo(s.T())
	s.accountCache = newMockAccountCache(s.T())
	s.emailCache = newMockEmailCache(s.T())
	s.passwordCache = newMockPasswordCache(s.T())
	s.sessionsCache = newMockSessionsCache(s.T())
	s.passManager = newMockPasswordManager(s.T())
	s.messenger = newMockMessenger(s.T())

	s.svc = New(ServiceDeps{
		Auth:          s.auth,
		AccountRepo:   s.accountRepo,
		EmailRepo:     s.emailRepo,
		PasswordRepo:  s.passwordRepo,
		SessionRepo:   s.sessionRepo,
		Tx:            &fakeTx{},
		AccountCache:  s.accountCache,
		EmailCache:    s.emailCache,
		PasswordCache: s.passwordCache,
		SessionsCache: s.sessionsCache,
		PassManager:   s.passManager,
		Messenger:     s.messenger,
	})
}

func TestAccountService(t *testing.T) {
	suite.Run(t, new(AccountServiceSuite))
}

// ─── Registration ───────────────────────────────────────────────────────────

func (s *AccountServiceSuite) TestRegistration_HappyPath() {
	ctx := context.Background()
	accountID := uuid.New()

	params := RegistrationParams{
		Email:    "user@example.com",
		Password: "Password1!",
		Role:     "user",
	}

	account := models.Account{ID: accountID}
	email := models.AccountEmail{AccountID: accountID, Email: params.Email}
	password := models.AccountPassword{AccountID: accountID, Hash: "hash"}

	s.passManager.On("GenerateHash", params.Password).Return("hash", nil)
	s.accountRepo.On("Create", mock.Anything, params).Return(account, nil)
	s.emailRepo.On("Create", mock.Anything, models.AccountEmail{AccountID: accountID, Email: params.Email}).Return(email, nil)
	s.passwordRepo.On("Create", mock.Anything, models.AccountPassword{AccountID: accountID, Hash: "hash"}).Return(password, nil)
	s.messenger.On("WriteAccountCreated", mock.Anything, account, email).Return(nil)
	s.accountCache.On("Set", mock.Anything, account).Return(nil).Maybe()
	s.emailCache.On("Set", mock.Anything, email).Return(nil).Maybe()
	s.passwordCache.On("Set", mock.Anything, password).Return(nil).Maybe()

	got, err := s.svc.Registration(ctx, params)

	require.NoError(s.T(), err)
	assert.Equal(s.T(), account, got)
}

func (s *AccountServiceSuite) TestRegistration_InvalidRole() {
	_, err := s.svc.Registration(context.Background(), RegistrationParams{
		Role:     "superadmin",
		Password: "Password1!",
	})

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, errx.ErrorRoleNotSupported)
}

func (s *AccountServiceSuite) TestRegistration_InvalidPassword_TooShort() {
	_, err := s.svc.Registration(context.Background(), RegistrationParams{
		Role:     "user",
		Password: "Ab1!",
	})

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, errx.ErrorPasswordIsNotAllowed)
}

func (s *AccountServiceSuite) TestRegistration_InvalidPassword_MissingSpecial() {
	_, err := s.svc.Registration(context.Background(), RegistrationParams{
		Role:     "user",
		Password: "Password1",
	})

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, errx.ErrorPasswordIsNotAllowed)
}

func (s *AccountServiceSuite) TestRegistration_GenerateHashError() {
	hashErr := errors.New("bcrypt error")
	s.passManager.On("GenerateHash", mock.Anything).Return("", hashErr)

	_, err := s.svc.Registration(context.Background(), RegistrationParams{
		Role:     "user",
		Password: "Password1!",
	})

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, hashErr)
}

func (s *AccountServiceSuite) TestRegistration_AccountRepoError() {
	repoErr := errors.New("db error")
	s.passManager.On("GenerateHash", mock.Anything).Return("hash", nil)
	s.accountRepo.On("Create", mock.Anything, mock.Anything).Return(models.Account{}, repoErr)

	_, err := s.svc.Registration(context.Background(), RegistrationParams{
		Role:     "user",
		Password: "Password1!",
	})

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, repoErr)
}

func (s *AccountServiceSuite) TestRegistration_EmailRepoError() {
	repoErr := errors.New("db error")
	account := models.Account{ID: uuid.New()}
	s.passManager.On("GenerateHash", mock.Anything).Return("hash", nil)
	s.accountRepo.On("Create", mock.Anything, mock.Anything).Return(account, nil)
	s.emailRepo.On("Create", mock.Anything, mock.Anything).Return(models.AccountEmail{}, repoErr)

	_, err := s.svc.Registration(context.Background(), RegistrationParams{
		Role:     "user",
		Password: "Password1!",
	})

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, repoErr)
}

func (s *AccountServiceSuite) TestRegistration_PasswordRepoError() {
	repoErr := errors.New("db error")
	account := models.Account{ID: uuid.New()}
	email := models.AccountEmail{AccountID: account.ID}
	s.passManager.On("GenerateHash", mock.Anything).Return("hash", nil)
	s.accountRepo.On("Create", mock.Anything, mock.Anything).Return(account, nil)
	s.emailRepo.On("Create", mock.Anything, mock.Anything).Return(email, nil)
	s.passwordRepo.On("Create", mock.Anything, mock.Anything).Return(models.AccountPassword{}, repoErr)

	_, err := s.svc.Registration(context.Background(), RegistrationParams{
		Role:     "user",
		Password: "Password1!",
	})

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, repoErr)
}

func (s *AccountServiceSuite) TestRegistration_MessengerError() {
	msgErr := errors.New("kafka error")
	account := models.Account{ID: uuid.New()}
	email := models.AccountEmail{AccountID: account.ID}
	password := models.AccountPassword{AccountID: account.ID}
	s.passManager.On("GenerateHash", mock.Anything).Return("hash", nil)
	s.accountRepo.On("Create", mock.Anything, mock.Anything).Return(account, nil)
	s.emailRepo.On("Create", mock.Anything, mock.Anything).Return(email, nil)
	s.passwordRepo.On("Create", mock.Anything, mock.Anything).Return(password, nil)
	s.messenger.On("WriteAccountCreated", mock.Anything, account, email).Return(msgErr)

	_, err := s.svc.Registration(context.Background(), RegistrationParams{
		Role:     "user",
		Password: "Password1!",
	})

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, msgErr)
}

// ─── GetMyAccountByID ────────────────────────────────────────────────────────

func (s *AccountServiceSuite) TestGetMyAccountByID_CacheHit() {
	accountID := uuid.New()
	account := models.Account{ID: accountID}

	s.accountCache.On("Get", mock.Anything, accountID).Return(account, nil)

	got, err := s.svc.GetMyAccountByID(context.Background(), models.AccountActor{ID: accountID})

	require.NoError(s.T(), err)
	assert.Equal(s.T(), account, got)
}

func (s *AccountServiceSuite) TestGetMyAccountByID_CacheMiss_RepoSuccess() {
	accountID := uuid.New()
	account := models.Account{ID: accountID}

	s.accountCache.On("Get", mock.Anything, accountID).Return(models.Account{}, errors.New("miss"))
	s.accountRepo.On("GetByID", mock.Anything, accountID).Return(account, nil)
	s.accountCache.On("Set", mock.Anything, account).Return(nil).Maybe()

	got, err := s.svc.GetMyAccountByID(context.Background(), models.AccountActor{ID: accountID})

	require.NoError(s.T(), err)
	assert.Equal(s.T(), account, got)
}

func (s *AccountServiceSuite) TestGetMyAccountByID_CacheMiss_RepoError() {
	accountID := uuid.New()
	repoErr := errors.New("db error")

	s.accountCache.On("Get", mock.Anything, accountID).Return(models.Account{}, errors.New("miss"))
	s.accountRepo.On("GetByID", mock.Anything, accountID).Return(models.Account{}, repoErr)

	_, err := s.svc.GetMyAccountByID(context.Background(), models.AccountActor{ID: accountID})

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, repoErr)
}

// ─── GetMyEmailByID ──────────────────────────────────────────────────────────

func (s *AccountServiceSuite) TestGetMyEmailByID_CacheHit() {
	accountID := uuid.New()
	email := models.AccountEmail{AccountID: accountID}

	s.emailCache.On("GetByID", mock.Anything, accountID).Return(email, nil)

	got, err := s.svc.GetMyEmailByID(context.Background(), models.AccountActor{ID: accountID})

	require.NoError(s.T(), err)
	assert.Equal(s.T(), email, got)
}

func (s *AccountServiceSuite) TestGetMyEmailByID_CacheMiss_RepoSuccess() {
	accountID := uuid.New()
	email := models.AccountEmail{AccountID: accountID}

	s.emailCache.On("GetByID", mock.Anything, accountID).Return(models.AccountEmail{}, errors.New("miss"))
	s.emailRepo.On("GetByID", mock.Anything, accountID).Return(email, nil)
	s.emailCache.On("Set", mock.Anything, email).Return(nil).Maybe()

	got, err := s.svc.GetMyEmailByID(context.Background(), models.AccountActor{ID: accountID})

	require.NoError(s.T(), err)
	assert.Equal(s.T(), email, got)
}

func (s *AccountServiceSuite) TestGetMyEmailByID_CacheMiss_RepoError() {
	accountID := uuid.New()
	repoErr := errors.New("db error")

	s.emailCache.On("GetByID", mock.Anything, accountID).Return(models.AccountEmail{}, errors.New("miss"))
	s.emailRepo.On("GetByID", mock.Anything, accountID).Return(models.AccountEmail{}, repoErr)

	_, err := s.svc.GetMyEmailByID(context.Background(), models.AccountActor{ID: accountID})

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, repoErr)
}

// ─── UpdatePassword ──────────────────────────────────────────────────────────

func (s *AccountServiceSuite) TestUpdatePassword_ValidateSessionError() {
	authErr := errors.New("session invalid")
	actor := models.AccountActor{ID: uuid.New(), SessionID: uuid.New()}

	s.auth.On("ValidateSession", mock.Anything, actor).Return(models.Account{}, models.Session{}, authErr)

	err := s.svc.UpdatePassword(context.Background(), actor, "old", "new")

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, authErr)
}

func (s *AccountServiceSuite) TestUpdatePassword_PasswordFromCache() {
	actor := models.AccountActor{ID: uuid.New(), SessionID: uuid.New()}
	pwd := models.AccountPassword{AccountID: actor.ID, Hash: "oldhash"}
	updated := models.AccountPassword{AccountID: actor.ID, Hash: "newhash"}

	s.auth.On("ValidateSession", mock.Anything, actor).Return(models.Account{}, models.Session{}, nil)
	s.passwordCache.On("Get", mock.Anything, actor.ID).Return(pwd, nil)
	s.passManager.On("CheckMatch", "OldPass1!", pwd.Hash).Return(nil)
	s.passManager.On("GenerateHash", "NewPass1!").Return("newhash", nil)
	s.passwordRepo.On("UpdatePassword", mock.Anything, actor.ID, "newhash").Return(updated, nil)
	s.passwordCache.On("Set", mock.Anything, updated).Return(nil).Maybe()

	err := s.svc.UpdatePassword(context.Background(), actor, "OldPass1!", "NewPass1!")

	require.NoError(s.T(), err)
}

func (s *AccountServiceSuite) TestUpdatePassword_PasswordFromRepo() {
	actor := models.AccountActor{ID: uuid.New(), SessionID: uuid.New()}
	pwd := models.AccountPassword{AccountID: actor.ID, Hash: "oldhash"}
	updated := models.AccountPassword{AccountID: actor.ID, Hash: "newhash"}

	s.auth.On("ValidateSession", mock.Anything, actor).Return(models.Account{}, models.Session{}, nil)
	s.passwordCache.On("Get", mock.Anything, actor.ID).Return(models.AccountPassword{}, errors.New("miss"))
	s.passwordRepo.On("GetByID", mock.Anything, actor.ID).Return(pwd, nil)
	s.passManager.On("CheckMatch", "OldPass1!", pwd.Hash).Return(nil)
	s.passManager.On("GenerateHash", "NewPass1!").Return("newhash", nil)
	s.passwordRepo.On("UpdatePassword", mock.Anything, actor.ID, "newhash").Return(updated, nil)
	s.passwordCache.On("Set", mock.Anything, updated).Return(nil).Maybe()

	err := s.svc.UpdatePassword(context.Background(), actor, "OldPass1!", "NewPass1!")

	require.NoError(s.T(), err)
}

func (s *AccountServiceSuite) TestUpdatePassword_RepoGetError() {
	actor := models.AccountActor{ID: uuid.New(), SessionID: uuid.New()}
	repoErr := errors.New("db error")

	s.auth.On("ValidateSession", mock.Anything, actor).Return(models.Account{}, models.Session{}, nil)
	s.passwordCache.On("Get", mock.Anything, actor.ID).Return(models.AccountPassword{}, errors.New("miss"))
	s.passwordRepo.On("GetByID", mock.Anything, actor.ID).Return(models.AccountPassword{}, repoErr)

	err := s.svc.UpdatePassword(context.Background(), actor, "OldPass1!", "NewPass1!")

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, repoErr)
}

func (s *AccountServiceSuite) TestUpdatePassword_WrongOldPassword() {
	actor := models.AccountActor{ID: uuid.New(), SessionID: uuid.New()}
	pwd := models.AccountPassword{AccountID: actor.ID, Hash: "oldhash"}
	checkErr := errors.New("password mismatch")

	s.auth.On("ValidateSession", mock.Anything, actor).Return(models.Account{}, models.Session{}, nil)
	s.passwordCache.On("Get", mock.Anything, actor.ID).Return(pwd, nil)
	s.passManager.On("CheckMatch", "WrongPass1!", pwd.Hash).Return(checkErr)

	err := s.svc.UpdatePassword(context.Background(), actor, "WrongPass1!", "NewPass1!")

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, checkErr)
}

func (s *AccountServiceSuite) TestUpdatePassword_InvalidNewPassword() {
	actor := models.AccountActor{ID: uuid.New(), SessionID: uuid.New()}
	pwd := models.AccountPassword{AccountID: actor.ID, Hash: "oldhash"}

	s.auth.On("ValidateSession", mock.Anything, actor).Return(models.Account{}, models.Session{}, nil)
	s.passwordCache.On("Get", mock.Anything, actor.ID).Return(pwd, nil)
	s.passManager.On("CheckMatch", "OldPass1!", pwd.Hash).Return(nil)

	err := s.svc.UpdatePassword(context.Background(), actor, "OldPass1!", "short")

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, errx.ErrorPasswordIsNotAllowed)
}

func (s *AccountServiceSuite) TestUpdatePassword_GenerateHashError() {
	actor := models.AccountActor{ID: uuid.New(), SessionID: uuid.New()}
	pwd := models.AccountPassword{AccountID: actor.ID, Hash: "oldhash"}
	hashErr := errors.New("bcrypt error")

	s.auth.On("ValidateSession", mock.Anything, actor).Return(models.Account{}, models.Session{}, nil)
	s.passwordCache.On("Get", mock.Anything, actor.ID).Return(pwd, nil)
	s.passManager.On("CheckMatch", "OldPass1!", pwd.Hash).Return(nil)
	s.passManager.On("GenerateHash", "NewPass1!").Return("", hashErr)

	err := s.svc.UpdatePassword(context.Background(), actor, "OldPass1!", "NewPass1!")

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, hashErr)
}

func (s *AccountServiceSuite) TestUpdatePassword_UpdateRepoError() {
	actor := models.AccountActor{ID: uuid.New(), SessionID: uuid.New()}
	pwd := models.AccountPassword{AccountID: actor.ID, Hash: "oldhash"}
	repoErr := errors.New("db error")

	s.auth.On("ValidateSession", mock.Anything, actor).Return(models.Account{}, models.Session{}, nil)
	s.passwordCache.On("Get", mock.Anything, actor.ID).Return(pwd, nil)
	s.passManager.On("CheckMatch", "OldPass1!", pwd.Hash).Return(nil)
	s.passManager.On("GenerateHash", "NewPass1!").Return("newhash", nil)
	s.passwordRepo.On("UpdatePassword", mock.Anything, actor.ID, "newhash").Return(models.AccountPassword{}, repoErr)

	err := s.svc.UpdatePassword(context.Background(), actor, "OldPass1!", "NewPass1!")

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, repoErr)
}

// ─── DeleteMyAccount ─────────────────────────────────────────────────────────

func (s *AccountServiceSuite) TestDeleteMyAccount_ValidateSessionError() {
	actor := models.AccountActor{ID: uuid.New(), SessionID: uuid.New()}
	authErr := errors.New("session invalid")

	s.auth.On("ValidateSession", mock.Anything, actor).Return(models.Account{}, models.Session{}, authErr)

	err := s.svc.DeleteMyAccount(context.Background(), actor)

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, authErr)
}

func (s *AccountServiceSuite) TestDeleteMyAccount_AccountRepoError() {
	actor := models.AccountActor{ID: uuid.New(), SessionID: uuid.New()}
	repoErr := errors.New("db error")

	s.auth.On("ValidateSession", mock.Anything, actor).Return(models.Account{}, models.Session{}, nil)
	s.accountRepo.On("Delete", mock.Anything, actor.ID).Return(models.Account{}, repoErr)

	err := s.svc.DeleteMyAccount(context.Background(), actor)

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, repoErr)
}

func (s *AccountServiceSuite) TestDeleteMyAccount_EmailRepoError() {
	actor := models.AccountActor{ID: uuid.New(), SessionID: uuid.New()}
	account := models.Account{ID: actor.ID}
	repoErr := errors.New("db error")

	s.auth.On("ValidateSession", mock.Anything, actor).Return(models.Account{}, models.Session{}, nil)
	s.accountRepo.On("Delete", mock.Anything, actor.ID).Return(account, nil)
	s.emailRepo.On("GetByID", mock.Anything, actor.ID, mock.Anything).Return(models.AccountEmail{}, repoErr)

	err := s.svc.DeleteMyAccount(context.Background(), actor)

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, repoErr)
}

func (s *AccountServiceSuite) TestDeleteMyAccount_SessionRepoError() {
	actor := models.AccountActor{ID: uuid.New(), SessionID: uuid.New()}
	account := models.Account{ID: actor.ID}
	email := models.AccountEmail{AccountID: actor.ID}
	repoErr := errors.New("db error")

	s.auth.On("ValidateSession", mock.Anything, actor).Return(models.Account{}, models.Session{}, nil)
	s.accountRepo.On("Delete", mock.Anything, actor.ID).Return(account, nil)
	s.emailRepo.On("GetByID", mock.Anything, actor.ID, mock.Anything).Return(email, nil)
	s.sessionRepo.On("DeleteManyForAccount", mock.Anything, actor.ID).Return([]uuid.UUID(nil), repoErr)

	err := s.svc.DeleteMyAccount(context.Background(), actor)

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, repoErr)
}

func (s *AccountServiceSuite) TestDeleteMyAccount_MessengerError() {
	actor := models.AccountActor{ID: uuid.New(), SessionID: uuid.New()}
	account := models.Account{ID: actor.ID}
	email := models.AccountEmail{AccountID: actor.ID}
	msgErr := errors.New("kafka error")

	s.auth.On("ValidateSession", mock.Anything, actor).Return(models.Account{}, models.Session{}, nil)
	s.accountRepo.On("Delete", mock.Anything, actor.ID).Return(account, nil)
	s.emailRepo.On("GetByID", mock.Anything, actor.ID, mock.Anything).Return(email, nil)
	s.sessionRepo.On("DeleteManyForAccount", mock.Anything, actor.ID).Return([]uuid.UUID{}, nil)
	s.messenger.On("WriteAccountDeleted", mock.Anything, account, email).Return(msgErr)

	err := s.svc.DeleteMyAccount(context.Background(), actor)

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, msgErr)
}

func (s *AccountServiceSuite) TestDeleteMyAccount_HappyPath() {
	actor := models.AccountActor{ID: uuid.New(), SessionID: uuid.New()}
	account := models.Account{ID: actor.ID}
	email := models.AccountEmail{AccountID: actor.ID}
	sessionID := uuid.New()

	s.auth.On("ValidateSession", mock.Anything, actor).Return(models.Account{}, models.Session{}, nil)
	s.accountRepo.On("Delete", mock.Anything, actor.ID).Return(account, nil)
	s.emailRepo.On("GetByID", mock.Anything, actor.ID, mock.Anything).Return(email, nil)
	s.sessionRepo.On("DeleteManyForAccount", mock.Anything, actor.ID).Return([]uuid.UUID{sessionID}, nil)
	s.messenger.On("WriteAccountDeleted", mock.Anything, account, email).Return(nil)
	s.accountCache.On("Delete", mock.Anything, actor.ID).Return(nil).Maybe()
	s.emailCache.On("DeleteByID", mock.Anything, actor.ID).Return(nil).Maybe()
	s.passwordCache.On("Delete", mock.Anything, actor.ID).Return(nil).Maybe()
	s.sessionsCache.On("Delete", mock.Anything, sessionID).Return(nil).Maybe()

	err := s.svc.DeleteMyAccount(context.Background(), actor)

	require.NoError(s.T(), err)
}
