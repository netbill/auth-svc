package rest

import (
	"context"
	"fmt"

	"github.com/netbill/restkit/auth"
)

type ctxKey int

const (
	AccountDataCtxKey ctxKey = iota
)

func AccountData(ctx context.Context) (auth.AccountData, error) {
	if ctx == nil {
		return auth.AccountData{}, fmt.Errorf("missing context")
	}

	userData, ok := ctx.Value(AccountDataCtxKey).(auth.AccountData)
	if !ok {
		return auth.AccountData{}, fmt.Errorf("missing context")
	}

	return userData, nil
}
