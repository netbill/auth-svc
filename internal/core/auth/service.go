package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/restkit/tokens"
)

type tokenManager interface {
	ParseAccountAuthAccess(tokenStr string) (tokens.AccountAuthClaims, error)
	ParseAccountAuthRefresh(enc string) (tokens.AccountAuthClaims, error)

	HashRefresh(rawRefresh string) (string, error)

	GenerateAccess(
		account models.Account, sessionID uuid.UUID,
	) (string, error)

	GenerateRefresh(
		account models.Account, sessionID uuid.UUID,
	) (string, error)
}

type messenger interface {
	WriteAccountCreated(ctx context.Context, account models.Account) error
	WriteAccountUsernameUpdated(ctx context.Context, account models.Account) error
	WriteAccountDeleted(ctx context.Context, accountID uuid.UUID) error
}

type PasswordManager interface {
	CheckRequirements(password string) error
	CheckPasswordMatch(password, hash string) error
	GenerateHash(password string) (string, error)
}

type Service struct {
	account   accountRepo
	email     emailRepo
	password  passwordRepo
	session   sessionRepo
	org       orgRepo
	tombstone tombstoneRepo
	tx        transaction

	token       tokenManager
	messenger   messenger
	passManager PasswordManager
}

type ServiceDeps struct {
	Account   accountRepo
	Email     emailRepo
	Password  passwordRepo
	Session   sessionRepo
	Org       orgRepo
	Tombstone tombstoneRepo
	Tx        transaction

	Token       tokenManager
	Messenger   messenger
	PassManager PasswordManager
}

func NewService(deps ServiceDeps) *Service {
	return &Service{
		account:     deps.Account,
		email:       deps.Email,
		password:    deps.Password,
		session:     deps.Session,
		org:         deps.Org,
		tombstone:   deps.Tombstone,
		tx:          deps.Tx,
		token:       deps.Token,
		messenger:   deps.Messenger,
		passManager: deps.PassManager,
	}
}
