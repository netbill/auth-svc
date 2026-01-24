package middlewares

import (
	"context"
	"fmt"

	"github.com/netbill/restkit/tokens"
)

const (
	accountDataCtxKey = iota
)

func AccountData(ctx context.Context) (tokens.AccountJwtData, error) {
	if ctx == nil {
		return tokens.AccountJwtData{}, fmt.Errorf("missing context")
	}

	userData, ok := ctx.Value(accountDataCtxKey).(tokens.AccountJwtData)
	if !ok {
		return tokens.AccountJwtData{}, fmt.Errorf("missing context")
	}

	return userData, nil
}
