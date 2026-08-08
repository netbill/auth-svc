package user

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/restkit/pagi"
	"github.com/netbill/restkit/tokens"
)

type DeletedFilter uint8

const (
	DeletedFilterActive  DeletedFilter = iota // 0 — default: only active (deleted_at IS NULL)
	DeletedFilterAll                          // active + deleted (no filter)
	DeletedFilterDeleted                      // only deleted (deleted_at IS NOT NULL)
)

type GetUserOptions struct {
	Deleted DeletedFilter
}

type GetUserOption func(*GetUserOptions)

func ApplyGetUserOptions(optFns []GetUserOption) GetUserOptions {
	var opts GetUserOptions
	for _, fn := range optFns {
		fn(&opts)
	}
	return opts
}

func WithDeleted(f DeletedFilter) GetUserOption {
	return func(o *GetUserOptions) {
		o.Deleted = f
	}
}

//go:generate mockery --name=auth --inpackage
type auth interface {
	ValidateSession(ctx context.Context, actor models.UserActor) (models.User, models.Session, error)
}

type Service struct {
	auth auth

	userRepo     userRepo
	emailRepo    emailRepo
	passwordRepo passwordRepo
	sessionRepo  sessionRepo

	tx transaction

	userCache     userCache
	emailCache    emailCache
	passwordCache passwordCache
	sessionsCache sessionsCache

	passManager passwordManager

	messenger messenger

	bucket   media
	username usernameGenerator
}

type ServiceDeps struct {
	Auth auth

	UserRepo     userRepo
	EmailRepo    emailRepo
	PasswordRepo passwordRepo
	SessionRepo  sessionRepo

	Tx transaction

	UserCache     userCache
	EmailCache    emailCache
	PasswordCache passwordCache
	SessionsCache sessionsCache

	PassManager passwordManager

	Messenger messenger

	Bucket   media
	Username usernameGenerator
}

func New(deps ServiceDeps) *Service {
	return &Service{
		auth:          deps.Auth,
		userRepo:      deps.UserRepo,
		emailRepo:     deps.EmailRepo,
		passwordRepo:  deps.PasswordRepo,
		sessionRepo:   deps.SessionRepo,
		tx:            deps.Tx,
		userCache:     deps.UserCache,
		emailCache:    deps.EmailCache,
		passwordCache: deps.PasswordCache,
		sessionsCache: deps.SessionsCache,
		passManager:   deps.PassManager,
		messenger:     deps.Messenger,
		bucket:        deps.Bucket,
		username:      deps.Username,
	}
}

//go:generate mockery --name=passwordManager --inpackage
type passwordManager interface {
	CheckMatch(password, hash string) error
	GenerateHash(password string) (string, error)
}

//go:generate mockery --name=messenger --inpackage
type messenger interface {
	WriteUserCreated(ctx context.Context, user models.User, email models.UserEmail) error
	WriteUserUpdated(ctx context.Context, user models.User) error
	WriteUserDeleted(ctx context.Context, user models.User, email models.UserEmail) error
}

type RegistrationParams struct {
	Email    string
	Password string
	Role     string
}

func (s *Service) Registration(
	ctx context.Context,
	params RegistrationParams,
) (models.User, error) {
	if err := tokens.ValidateUserSystemRole(params.Role); err != nil {
		return models.User{}, errx.ErrorRoleNotSupported.Raise(err)
	}

	if err := s.checkPasswordRequirements(params.Password); err != nil {
		return models.User{}, err
	}

	hash, err := s.passManager.GenerateHash(params.Password)
	if err != nil {
		return models.User{}, err
	}

	candidate := s.username.Generate()

	var user models.User
	var email models.UserEmail
	var password models.UserPassword

	if err = s.tx.Transaction(ctx, func(ctx context.Context) error {
		user, err = s.userRepo.Create(ctx, params, candidate)
		if err != nil {
			return err
		}

		email, err = s.emailRepo.Create(ctx, models.UserEmail{
			UserID: user.ID,
			Email:  params.Email,
		})
		if err != nil {
			return err
		}

		password, err = s.passwordRepo.Create(ctx, models.UserPassword{
			UserID: user.ID,
			Hash:   hash,
		})
		if err != nil {
			return err
		}

		return s.messenger.WriteUserCreated(ctx, user, email)
	}); err != nil {
		return models.User{}, err
	}

	detached := context.WithoutCancel(ctx)
	go s.userCache.Set(detached, user)
	go s.emailCache.Set(detached, email)
	go s.passwordCache.Set(detached, password)

	return user, nil
}

func (s *Service) GetMyUserByID(
	ctx context.Context,
	actor models.UserActor,
) (models.User, error) {
	if cached, err := s.userCache.Get(ctx, actor.ID); err == nil {
		return cached, nil
	}

	user, err := s.userRepo.GetByID(ctx, actor.ID)
	if err != nil {
		return models.User{}, err
	}

	go s.userCache.Set(context.WithoutCancel(ctx), user)

	return user, nil
}

func (s *Service) GetMyEmailByID(
	ctx context.Context,
	actor models.UserActor,
) (models.UserEmail, error) {
	if cached, err := s.emailCache.GetByID(ctx, actor.ID); err == nil {
		return cached, nil
	}

	email, err := s.emailRepo.GetByID(ctx, actor.ID)
	if err != nil {
		return models.UserEmail{}, err
	}

	go s.emailCache.Set(context.WithoutCancel(ctx), email)

	return email, nil
}

// GetUserByID is the public (non-"me") lookup used for profile pages —
// unlike GetMyUserByID it is not cached under the actor's session.
func (s *Service) GetUserByID(
	ctx context.Context,
	userID uuid.UUID,
) (models.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *Service) GetByUsername(
	ctx context.Context,
	username string,
) (models.User, error) {
	return s.userRepo.GetByUsername(ctx, username)
}

type FilterParams struct {
	Text *string
}

func (s *Service) GetList(
	ctx context.Context,
	params FilterParams,
	limit, offset uint,
) (pagi.Page[[]models.User], error) {
	return s.userRepo.Filter(ctx, params, limit, offset)
}

type UpdateParams struct {
	AvatarKey   *string
	Pseudonym   *string
	Description *string
}

func (p UpdateParams) HasChanges(m models.User) bool {
	return !ptrEqual(p.AvatarKey, m.AvatarKey) ||
		!ptrEqual(p.Pseudonym, m.Pseudonym) ||
		!ptrEqual(p.Description, m.Description)
}

func ptrEqual[T comparable](a, b *T) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func (s *Service) Update(
	ctx context.Context,
	actor models.UserActor,
	params UpdateParams,
) (u models.User, err error) {
	u, err = s.userRepo.GetByID(ctx, actor.ID)
	if err != nil {
		return models.User{}, err
	}

	if !params.HasChanges(u) {
		return u, nil
	}

	switch {
	case params.AvatarKey != nil && *params.AvatarKey == "" && u.AvatarKey != nil:
		if err := s.bucket.DeleteUserAvatar(ctx, actor.ID, *u.AvatarKey); err != nil {
			return models.User{}, fmt.Errorf("failed to delete user avatar: %w", err)
		}
		params.AvatarKey = nil
	case params.AvatarKey != nil:
		avatarKey, err := s.bucket.UpdateUserAvatar(ctx, actor.ID, *params.AvatarKey)
		if err != nil {
			return models.User{}, fmt.Errorf("failed to update user avatar: %w", err)
		}
		params.AvatarKey = &avatarKey
	}

	if err = s.tx.Transaction(ctx, func(ctx context.Context) error {
		u, err = s.userRepo.Update(ctx, actor.ID, params)
		if err != nil {
			return err
		}

		return s.messenger.WriteUserUpdated(ctx, u)
	}); err != nil {
		return models.User{}, err
	}

	go s.userCache.Set(context.WithoutCancel(ctx), u)

	return u, nil
}

func (s *Service) UpdateUsername(
	ctx context.Context,
	actor models.UserActor,
	newUsername string,
) (models.User, error) {
	if err := s.username.Validate(newUsername); err != nil {
		return models.User{}, err
	}

	current, err := s.userRepo.GetByID(ctx, actor.ID)
	if err != nil {
		return models.User{}, err
	}

	if current.Username == newUsername {
		return current, nil
	}

	unavailable, err := s.userRepo.ExistByUsername(ctx, newUsername)
	if err != nil {
		return models.User{}, err
	}
	if unavailable {
		return models.User{}, errx.ErrorUsernameTaken
	}

	var u models.User
	if err = s.tx.Transaction(ctx, func(ctx context.Context) error {
		u, err = s.userRepo.UpdateUsername(ctx, actor.ID, newUsername)
		if err != nil {
			return err
		}

		return s.messenger.WriteUserUpdated(ctx, u)
	}); err != nil {
		return models.User{}, err
	}

	go s.userCache.Set(context.WithoutCancel(ctx), u)

	return u, nil
}

func (s *Service) CreateUploadMediaLinks(
	ctx context.Context,
	actor models.UserActor,
) (models.User, models.UploadUserMediaLinks, error) {
	u, err := s.userRepo.GetByID(ctx, actor.ID)
	if err != nil {
		return models.User{}, models.UploadUserMediaLinks{}, err
	}

	links, err := s.bucket.CreateUserAvatarUploadMediaLinks(ctx, actor.ID)
	if err != nil {
		return models.User{}, models.UploadUserMediaLinks{}, err
	}

	return u, models.UploadUserMediaLinks{
		Avatar: links,
	}, nil
}

type DeleteUploadMediaParams struct {
	Avatar *string
}

func (s *Service) DeleteUploadMedia(
	ctx context.Context,
	actor models.UserActor,
	params DeleteUploadMediaParams,
) error {
	if params.Avatar != nil {
		if err := s.bucket.DeleteUploadUserAvatar(ctx, actor.ID, *params.Avatar); err != nil {
			return fmt.Errorf("failed to delete upload user avatar: %w", err)
		}
	}

	return nil
}

func (s *Service) UpdatePassword(
	ctx context.Context,
	actor models.UserActor,
	oldPassword, newPassword string,
) error {
	if _, _, err := s.auth.ValidateSession(ctx, actor); err != nil {
		return err
	}

	pwd, err := s.passwordCache.Get(ctx, actor.ID)
	if err != nil {
		pwd, err = s.passwordRepo.GetByID(ctx, actor.ID)
		if err != nil {
			return err
		}
	}

	if err = s.passManager.CheckMatch(oldPassword, pwd.Hash); err != nil {
		return err
	}

	if err = s.checkPasswordRequirements(newPassword); err != nil {
		return err
	}

	hash, err := s.passManager.GenerateHash(newPassword)
	if err != nil {
		return err
	}

	var updated models.UserPassword
	if err = s.tx.Transaction(ctx, func(ctx context.Context) error {
		updated, err = s.passwordRepo.UpdatePassword(ctx, actor.ID, hash)
		return err
	}); err != nil {
		return err
	}

	go s.passwordCache.Set(context.WithoutCancel(ctx), updated)

	return nil
}

func (s *Service) DeleteMyUser(
	ctx context.Context,
	actor models.UserActor,
) error {
	if _, _, err := s.auth.ValidateSession(ctx, actor); err != nil {
		return err
	}

	var (
		user       models.User
		email      models.UserEmail
		sessionIDs []uuid.UUID
		err        error
	)

	if err = s.tx.Transaction(ctx, func(ctx context.Context) error {
		user, err = s.userRepo.Delete(ctx, actor.ID)
		if err != nil {
			return err
		}

		email, err = s.emailRepo.GetByID(ctx, actor.ID, WithDeleted(DeletedFilterAll))
		if err != nil {
			return err
		}

		sessionIDs, err = s.sessionRepo.DeleteManyForUser(ctx, actor.ID)
		if err != nil {
			return err
		}

		return s.messenger.WriteUserDeleted(ctx, user, email)
	}); err != nil {
		return err
	}

	detached := context.WithoutCancel(ctx)

	go s.userCache.Delete(detached, actor.ID)
	go s.emailCache.DeleteByID(detached, actor.ID)
	go s.passwordCache.Delete(detached, actor.ID)

	for _, id := range sessionIDs {
		go s.sessionsCache.Delete(detached, id)
	}

	return nil
}

func (s *Service) checkPasswordRequirements(password string) error {
	if len(password) < 8 || len(password) > 32 {
		return errx.ErrorPasswordIsNotAllowed.Raise(
			fmt.Errorf("password must be between 8 and 32 characters"),
		)
	}

	var (
		hasUpper, hasLower, hasDigit, hasSpecial bool
	)

	allowedSpecials := "-.!#$%&?,@"

	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case strings.ContainsRune(allowedSpecials, r):
			hasSpecial = true
		default:
			return errx.ErrorPasswordIsNotAllowed.Raise(
				fmt.Errorf("password contains invalid characters %s", string(r)),
			)
		}
	}

	if !hasUpper {
		return errx.ErrorPasswordIsNotAllowed.Raise(fmt.Errorf("need at least one uppercase letter"))
	}
	if !hasLower {
		return errx.ErrorPasswordIsNotAllowed.Raise(fmt.Errorf("need at least one lower case letter"))
	}
	if !hasDigit {
		return errx.ErrorPasswordIsNotAllowed.Raise(fmt.Errorf("need at least one digit"))
	}
	if !hasSpecial {
		return errx.ErrorPasswordIsNotAllowed.Raise(
			fmt.Errorf("need at least one special character from %s", allowedSpecials),
		)
	}

	return nil
}
