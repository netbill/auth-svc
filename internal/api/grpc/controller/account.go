package controller

import (
	"context"
	"errors"

	"github.com/netbill/auth-svc/internal/api/grpc/reponses"
	"github.com/netbill/auth-svc/internal/api/grpc/scope"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/internal/modules/account"
	"github.com/netbill/auth-svc/pkg/pb"
	"github.com/netbill/restkit/tokens"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type AccountCore interface {
	Registration(ctx context.Context, params account.RegistrationParams) (models.Account, error)
	GetMyAccountByID(ctx context.Context, actor models.AccountActor) (models.Account, error)
	GetMyEmailByID(ctx context.Context, actor models.AccountActor) (models.AccountEmail, error)
	UpdateUsername(ctx context.Context, actor models.AccountActor, newUsername string) (models.Account, error)
	UpdatePassword(ctx context.Context, actor models.AccountActor, oldPassword, newPassword string) error
	DeleteMyAccount(ctx context.Context, actor models.AccountActor) error
}

type AccountServer struct {
	pb.UnimplementedAccountServiceServer
	accounts AccountCore
}

func NewAccountServer(accounts AccountCore) *AccountServer {
	return &AccountServer{accounts: accounts}
}

const operationCreateAccount = "create_account"

func (s *AccountServer) CreateAccount(ctx context.Context, req *pb.CreateAccountRequest) (*pb.CreateAccountResponse, error) {
	log := scope.Log(ctx).WithOperation(operationCreateAccount)

	acc, err := s.accounts.Registration(ctx, account.RegistrationParams{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
		Role:     tokens.RoleSystemUser,
	})
	switch {
	case errors.Is(err, errx.ErrorEmailAlreadyExist):
		log.Warn("email already exists", "error", err)
		return nil, status.Error(codes.AlreadyExists, "email already exists")
	case errors.Is(err, errx.ErrorUsernameAlreadyTaken):
		log.Warn("username already taken", "error", err)
		return nil, status.Error(codes.AlreadyExists, "username already taken")
	case errors.Is(err, errx.ErrorUsernameIsNotAllowed):
		log.Warn("username is not allowed", "error", err)
		return nil, status.Error(codes.InvalidArgument, "username is not allowed")
	case errors.Is(err, errx.ErrorPasswordIsNotAllowed):
		log.Warn("password is not allowed", "error", err)
		return nil, status.Error(codes.InvalidArgument, "password is not allowed")
	case err != nil:
		log.Error("unexpected error", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	default:
		log.Info("account created")
		return &pb.CreateAccountResponse{Account: reponses.Account(acc)}, nil
	}
}

const operationGetMyAccount = "get_my_account"

func (s *AccountServer) GetMyAccount(ctx context.Context, _ *pb.GetMyAccountRequest) (*pb.GetMyAccountResponse, error) {
	log := scope.Log(ctx).WithOperation(operationGetMyAccount)

	acc, err := s.accounts.GetMyAccountByID(ctx, scope.AccountActor(ctx))
	switch {
	case errors.Is(err, errx.ErrorAccountNotFound),
		errors.Is(err, errx.ErrorAccountDeleted):
		log.Warn("account not found", "error", err)
		return nil, status.Error(codes.NotFound, "account not found")
	case err != nil:
		log.Error("unexpected error", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	default:
		log.Info("account retrieved")
		return &pb.GetMyAccountResponse{Account: reponses.Account(acc)}, nil
	}
}

const operationGetMyEmail = "get_my_email"

func (s *AccountServer) GetMyEmail(ctx context.Context, _ *pb.GetMyEmailRequest) (*pb.GetMyEmailResponse, error) {
	log := scope.Log(ctx).WithOperation(operationGetMyEmail)

	email, err := s.accounts.GetMyEmailByID(ctx, scope.AccountActor(ctx))
	switch {
	case errors.Is(err, errx.ErrorAccountNotFound):
		log.Warn("account not found", "error", err)
		return nil, status.Error(codes.NotFound, "account not found")
	case err != nil:
		log.Error("unexpected error", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	default:
		log.Info("email retrieved")
		return &pb.GetMyEmailResponse{Email: reponses.AccountEmail(email)}, nil
	}
}

const operationUpdateUsername = "update_username"

func (s *AccountServer) UpdateUsername(ctx context.Context, req *pb.UpdateUsernameRequest) (*pb.UpdateUsernameResponse, error) {
	log := scope.Log(ctx).WithOperation(operationUpdateUsername)

	acc, err := s.accounts.UpdateUsername(ctx, scope.AccountActor(ctx), req.NewUsername)
	switch {
	case errors.Is(err, errx.ErrorUsernameAlreadyTaken):
		log.Warn("username already taken", "error", err)
		return nil, status.Error(codes.AlreadyExists, "username already taken")
	case errors.Is(err, errx.ErrorUsernameIsNotAllowed):
		log.Warn("username is not allowed", "error", err)
		return nil, status.Error(codes.InvalidArgument, "username is not allowed")
	case errors.Is(err, errx.ErrorAccountInvalidSession):
		log.Warn("invalid session", "error", err)
		return nil, status.Error(codes.Unauthenticated, "invalid session")
	case err != nil:
		log.Error("unexpected error", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	default:
		log.Info("username updated")
		return &pb.UpdateUsernameResponse{Account: reponses.Account(acc)}, nil
	}
}

const operationUpdatePassword = "update_password"

func (s *AccountServer) UpdatePassword(ctx context.Context, req *pb.UpdatePasswordRequest) (*emptypb.Empty, error) {
	log := scope.Log(ctx).WithOperation(operationUpdatePassword)

	err := s.accounts.UpdatePassword(ctx, scope.AccountActor(ctx), req.OldPassword, req.NewPassword)
	switch {
	case errors.Is(err, errx.ErrorPasswordInvalid):
		log.Warn("invalid password", "error", err)
		return nil, status.Error(codes.Unauthenticated, "invalid password")
	case errors.Is(err, errx.ErrorPasswordIsNotAllowed):
		log.Warn("password is not allowed", "error", err)
		return nil, status.Error(codes.InvalidArgument, "password is not allowed")
	case errors.Is(err, errx.ErrorCannotChangePasswordYet):
		log.Warn("cannot change password yet", "error", err)
		return nil, status.Error(codes.FailedPrecondition, "cannot change password yet")
	case errors.Is(err, errx.ErrorAccountInvalidSession):
		log.Warn("invalid session", "error", err)
		return nil, status.Error(codes.Unauthenticated, "invalid session")
	case err != nil:
		log.Error("unexpected error", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	default:
		log.Info("password updated")
		return &emptypb.Empty{}, nil
	}
}

const operationDeleteMyAccount = "delete_my_account"

func (s *AccountServer) DeleteMyAccount(ctx context.Context, _ *pb.DeleteMyAccountRequest) (*emptypb.Empty, error) {
	log := scope.Log(ctx).WithOperation(operationDeleteMyAccount)

	err := s.accounts.DeleteMyAccount(ctx, scope.AccountActor(ctx))
	switch {
	case errors.Is(err, errx.ErrorAccountInvalidSession):
		log.Warn("invalid session", "error", err)
		return nil, status.Error(codes.Unauthenticated, "invalid session")
	case err != nil:
		log.Error("unexpected error", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	default:
		log.Info("account deleted")
		return &emptypb.Empty{}, nil
	}
}
