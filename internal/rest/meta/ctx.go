package meta

import (
	"context"
	"fmt"

	"github.com/netbill/restkit/token"
)

type ctxKey int

const (
	AccountDataCtxKey ctxKey = iota
)

func AccountData(ctx context.Context) (token.AccountData, error) {
	if ctx == nil {
		return token.AccountData{}, fmt.Errorf("missing context")
	}

	userData, ok := ctx.Value(AccountDataCtxKey).(token.AccountData)
	if !ok {
		return token.AccountData{}, fmt.Errorf("missing context")
	}

	return userData, nil
}
