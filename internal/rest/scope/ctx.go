package scope

import (
	"context"
	"net/http"

	"github.com/netbill/auth-svc/internal/core/models"
	"github.com/netbill/auth-svc/pkg/log"
	"github.com/netbill/restkit/tokens"
)

type ctxKey int

const (
	LogCtxKey ctxKey = iota
	AccountDataCtxKey
)

func CtxLog(ctx context.Context, log *log.Logger) context.Context {
	return context.WithValue(ctx, LogCtxKey, log)
}

func Log(r *http.Request) *log.Logger {
	log := r.Context().Value(LogCtxKey).(*log.Logger)

	authClaims, ok := r.Context().Value(AccountDataCtxKey).(tokens.AccountAuthClaims)
	if ok {
		log = log.WithAccountAuthClaims(authClaims)
	}

	return log
}

func CtxAccountAuth(ctx context.Context, account tokens.AccountAuthClaims) context.Context {
	return context.WithValue(ctx, AccountDataCtxKey, account)
}

func AccountActor(r *http.Request) models.AccountActor {
	claims := r.Context().Value(AccountDataCtxKey).(tokens.AccountAuthClaims)
	return models.AccountActor{
		ID:        claims.GetAccountID(),
		SessionID: claims.GetSessionID(),
		Role:      claims.GetRole(),
	}
}
