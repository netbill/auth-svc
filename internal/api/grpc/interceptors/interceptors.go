package interceptors

import (
	"context"
	"strings"

	"github.com/netbill/auth-svc/internal/api/grpc/scope"
	"github.com/netbill/auth-svc/pkg/log"
	"github.com/netbill/restkit/tokens"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var publicMethods = map[string]struct{}{
	"/auth.v1.AccountService/CreateAccount":   {},
	"/auth.v1.SessionService/LoginByEmail":    {},
	"/auth.v1.SessionService/LoginByUsername": {},
	"/auth.v1.SessionService/LoginByGoogle":   {},
	"/auth.v1.SessionService/Refresh":         {},
	"/auth.v1.AuthService/ValidateSession":    {},
}

type TokenParser interface {
	ParseAccountAuthAccess(tokenStr string) (tokens.AccountAuthClaims, error)
}

func LogInterceptor(logger *log.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ctx = scope.CtxWithLog(ctx, logger)

		return handler(ctx, req)
	}
}

func AuthInterceptor(tokenMgr TokenParser) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if _, ok := publicMethods[info.FullMethod]; ok {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		values := md["authorization"]
		if len(values) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}

		tokenStr := strings.TrimPrefix(values[0], "Bearer ")
		claims, err := tokenMgr.ParseAccountAuthAccess(tokenStr)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		ctx = scope.CtxWithClaims(ctx, claims)

		return handler(ctx, req)
	}
}
