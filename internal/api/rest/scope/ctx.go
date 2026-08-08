package scope

import (
	"context"
	"net/http"

	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/pkg/log"
	"github.com/netbill/restkit/tokens"
)

type ctxKey int

const (
	LogCtxKey ctxKey = iota
	UserDataCtxKey
)

func CtxLog(ctx context.Context, log *log.Logger) context.Context {
	return context.WithValue(ctx, LogCtxKey, log.With("api", "rest"))
}

func Log(r *http.Request) *log.Logger {
	log := r.Context().Value(LogCtxKey).(*log.Logger)

	authClaims, ok := r.Context().Value(UserDataCtxKey).(tokens.AccountAuthClaims)
	if ok {
		log = log.WithUserAuthClaims(authClaims)
	}

	return log
}

func CtxUserAuth(ctx context.Context, user tokens.AccountAuthClaims) context.Context {
	return context.WithValue(ctx, UserDataCtxKey, user)
}

func UserActor(r *http.Request) models.UserActor {
	claims := r.Context().Value(UserDataCtxKey).(tokens.AccountAuthClaims)
	return models.UserActor{
		ID:        claims.GetAccountID(),
		SessionID: claims.GetSessionID(),
		Role:      claims.GetRole(),
	}
}
