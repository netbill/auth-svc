package controller

import (
	"context"
	"errors"
	"strings"

	"github.com/netbill/auth-svc/internal/api/grpc/interceptors"
	"github.com/netbill/auth-svc/internal/api/grpc/reponses"
	"github.com/netbill/auth-svc/internal/api/grpc/scope"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type authCore interface {
	ValidateSession(ctx context.Context, actor models.AccountActor) (models.Account, models.Session, error)
}

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	auth     authCore
	tokenMgr interceptors.TokenParser
}

func NewAuthServer(auth authCore, tokenMgr interceptors.TokenParser) *AuthServer {
	return &AuthServer{auth: auth, tokenMgr: tokenMgr}
}

const operationValidateSession = "validate_session"

func (s *AuthServer) ValidateSession(ctx context.Context, _ *pb.ValidateSessionRequest) (*pb.ValidateSessionResponse, error) {
	log := scope.Log(ctx).WithOperation(operationValidateSession)

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	values := md["authorization"]
	if len(values) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing authorization header")
	}

	claims, err := s.tokenMgr.ParseAccountAuthAccess(strings.TrimPrefix(values[0], "Bearer "))
	if err != nil {
		log.Debug("validate session: invalid token", "error", err)
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	actor := models.AccountActor{
		ID:        claims.GetAccountID(),
		SessionID: claims.GetSessionID(),
		Role:      claims.GetRole(),
	}

	account, sess, err := s.auth.ValidateSession(ctx, actor)
	switch {
	case errors.Is(err, errx.ErrorAccountNotFound),
		errors.Is(err, errx.ErrorAccountDeleted),
		errors.Is(err, errx.ErrorSessionNotFound),
		errors.Is(err, errx.ErrorSessionDeleted),
		errors.Is(err, errx.ErrorAccountInvalidSession):
		log.Warn("validate session: invalid session", "error", err)
		return nil, status.Error(codes.Unauthenticated, "invalid session")
	case err != nil:
		log.Error("validate session: unexpected error", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	default:
		log.Info("validate session: session valid")
		return &pb.ValidateSessionResponse{
			Account: reponses.Account(account),
			Session: reponses.Session(sess),
		}, nil
	}
}
