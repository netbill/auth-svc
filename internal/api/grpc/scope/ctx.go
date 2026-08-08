package scope

import (
	"context"

	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/pkg/log"
	"github.com/netbill/restkit/tokens"
)

type ctxKey int

const (
	UserDataCtxKey ctxKey = iota
	LogCtxKey
)

func CtxWithClaims(ctx context.Context, claims tokens.AccountAuthClaims) context.Context {
	return context.WithValue(ctx, UserDataCtxKey, claims)
}

func UserActor(ctx context.Context) models.UserActor {
	claims := ctx.Value(UserDataCtxKey).(tokens.AccountAuthClaims)
	return models.UserActor{
		ID:        claims.GetAccountID(),
		SessionID: claims.GetSessionID(),
		Role:      claims.GetRole(),
	}
}

func CtxWithLog(ctx context.Context, log *log.Logger) context.Context {
	return context.WithValue(ctx, LogCtxKey, log.With("api", "grpc"))
}

func Log(ctx context.Context) *log.Logger {
	log := ctx.Value(LogCtxKey).(*log.Logger)

	authClaims, ok := ctx.Value(UserDataCtxKey).(tokens.AccountAuthClaims)
	if ok {
		log = log.WithUserAuthClaims(authClaims)
	}

	return log
}
