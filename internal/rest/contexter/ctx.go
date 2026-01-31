package contexter

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/netbill/restkit/tokens"
)

const (
	AccountDataCtxKey   = iota
	UploadContentCtxKey = iota
)

type Account interface {
	GetAccountID() uuid.UUID
	GetSessionID() uuid.UUID
	GetAccountRole() string
}

func AccountData(ctx context.Context) (Account, error) {
	if ctx == nil {
		return nil, fmt.Errorf("missing context")
	}

	userData, ok := ctx.Value(AccountDataCtxKey).(tokens.AccountClaims)
	if !ok {
		return nil, fmt.Errorf("missing context")
	}

	if err := userData.Validate(); err != nil {
		return nil, fmt.Errorf("invalid account data in context: %w", err)
	}

	return userData, nil
}

type UploadContent interface {
	GetOwnerAccountID() uuid.UUID
	GetUploadSessionID() uuid.UUID
	GetResourceID() string
	GetResource() string
}

func UploadContentData(ctx context.Context) (UploadContent, error) {
	if ctx == nil {
		return nil, fmt.Errorf("missing context")
	}

	userData, ok := ctx.Value(UploadContentCtxKey).(tokens.UploadContentClaims)
	if !ok {
		return nil, fmt.Errorf("missing context")
	}

	if err := userData.Validate(); err != nil {
		return nil, fmt.Errorf("invalid upload content data in context: %w", err)
	}

	return userData, nil
}
