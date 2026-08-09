package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/restkit/pagi"
)

//go:generate mockery --name=transaction --inpackage
type transaction interface {
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}

//go:generate mockery --name=userRepo --inpackage
type userRepo interface {
	Create(ctx context.Context, params RegistrationParams) (models.User, error)
	GetByID(ctx context.Context, userID uuid.UUID) (models.User, error)
	GetByUsername(ctx context.Context, username string) (models.User, error)
	ExistByUsername(ctx context.Context, username string) (bool, error)
	Update(ctx context.Context, userID uuid.UUID, params UpdateParams) (models.User, error)
	UpdateUsername(ctx context.Context, userID uuid.UUID, username string) (models.User, error)
	Filter(ctx context.Context, params FilterParams, limit, offset uint) (pagi.Page[[]models.User], error)
	Delete(ctx context.Context, userID uuid.UUID) (models.User, error)
}

//go:generate mockery --name=media --inpackage
type media interface {
	CreateUserAvatarUploadMediaLinks(
		ctx context.Context,
		userID uuid.UUID,
	) (models.UploadMediaLink, error)

	DeleteUploadUserAvatar(
		ctx context.Context,
		userID uuid.UUID,
		key string,
	) error

	DeleteUserAvatar(
		ctx context.Context,
		userID uuid.UUID,
		key string,
	) error

	UpdateUserAvatar(
		ctx context.Context,
		userID uuid.UUID,
		key string,
	) (string, error)
}

//go:generate mockery --name=usernameValidator --inpackage
type usernameValidator interface {
	Validate(username string) error
}

//go:generate mockery --name=emailRepo --inpackage
type emailRepo interface {
	Create(ctx context.Context, params models.UserEmail) (models.UserEmail, error)
	GetByID(ctx context.Context, userID uuid.UUID, opts ...GetUserOption) (models.UserEmail, error)
}

//go:generate mockery --name=sessionRepo --inpackage
type sessionRepo interface {
	DeleteManyForUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}

//go:generate mockery --name=passwordRepo --inpackage
type passwordRepo interface {
	Create(ctx context.Context, params models.UserPassword) (models.UserPassword, error)
	GetByID(ctx context.Context, userID uuid.UUID) (models.UserPassword, error)
	UpdatePassword(
		ctx context.Context,
		userID uuid.UUID,
		passwordHash string,
	) (models.UserPassword, error)
}
