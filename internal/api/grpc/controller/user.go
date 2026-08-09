package controller

import (
	"context"
	"errors"

	"github.com/netbill/auth-svc/internal/api/grpc/reponses"
	"github.com/netbill/auth-svc/internal/api/grpc/scope"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/internal/modules/user"
	"github.com/netbill/auth-svc/pkg/pb"
	"github.com/netbill/restkit/tokens"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UserCore interface {
	Registration(ctx context.Context, params user.RegistrationParams) (models.User, error)
	GetMyUserByID(ctx context.Context, actor models.UserActor) (models.User, error)
	GetMyEmailByID(ctx context.Context, actor models.UserActor) (models.UserEmail, error)
	UpdatePassword(ctx context.Context, actor models.UserActor, oldPassword, newPassword string) error
	DeleteMyUser(ctx context.Context, actor models.UserActor) error
}

type UserMetrics interface {
	RecordRegistration(ctx context.Context, err *error)
}

type UserServer struct {
	pb.UnimplementedUserServiceServer
	users   UserCore
	metrics UserMetrics
}

func NewUserServer(users UserCore, m UserMetrics) *UserServer {
	return &UserServer{users: users, metrics: m}
}

const operationCreateUser = "create_user"

func (s *UserServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (_ *pb.CreateUserResponse, err error) {
	log := scope.Log(ctx).WithOperation(operationCreateUser)
	defer s.metrics.RecordRegistration(ctx, &err)

	acc, err := s.users.Registration(ctx, user.RegistrationParams{
		Email:    req.Email,
		Password: req.Password,
		Username: req.Username,
		Role:     tokens.RoleSystemUser,
	})
	switch {
	case errors.Is(err, errx.ErrorEmailAlreadyExist):
		log.Warn("email already exists", "error", err)
		return nil, status.Error(codes.AlreadyExists, "email already exists")
	case errors.Is(err, errx.ErrorUsernameTaken):
		log.Warn("username already exists", "error", err)
		return nil, status.Error(codes.AlreadyExists, "username already exists")
	case errors.Is(err, errx.ErrorPasswordIsNotAllowed):
		log.Warn("password is not allowed", "error", err)
		return nil, status.Error(codes.InvalidArgument, "password is not allowed")
	case errors.Is(err, errx.ErrorUsernameNotValid):
		log.Warn("username is not valid", "error", err)
		return nil, status.Error(codes.InvalidArgument, "username is not valid")
	case err != nil:
		log.Error("unexpected error", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	default:
		log.Info("user created")
		return &pb.CreateUserResponse{User: reponses.User(acc)}, nil
	}
}

const operationGetMyUser = "get_my_user"

func (s *UserServer) GetMyUser(ctx context.Context, _ *pb.GetMyUserRequest) (*pb.GetMyUserResponse, error) {
	log := scope.Log(ctx).WithOperation(operationGetMyUser)

	acc, err := s.users.GetMyUserByID(ctx, scope.UserActor(ctx))
	switch {
	case errors.Is(err, errx.ErrorUserNotFound),
		errors.Is(err, errx.ErrorUserDeleted):
		log.Warn("user not found", "error", err)
		return nil, status.Error(codes.NotFound, "user not found")
	case err != nil:
		log.Error("unexpected error", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	default:
		log.Info("user retrieved")
		return &pb.GetMyUserResponse{User: reponses.User(acc)}, nil
	}
}

const operationGetMyEmail = "get_my_email"

func (s *UserServer) GetMyEmail(ctx context.Context, _ *pb.GetMyEmailRequest) (*pb.GetMyEmailResponse, error) {
	log := scope.Log(ctx).WithOperation(operationGetMyEmail)

	email, err := s.users.GetMyEmailByID(ctx, scope.UserActor(ctx))
	switch {
	case errors.Is(err, errx.ErrorUserNotFound):
		log.Warn("user not found", "error", err)
		return nil, status.Error(codes.NotFound, "user not found")
	case err != nil:
		log.Error("unexpected error", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	default:
		log.Info("email retrieved")
		return &pb.GetMyEmailResponse{Email: reponses.UserEmail(email)}, nil
	}
}

const operationUpdatePassword = "update_password"

func (s *UserServer) UpdatePassword(ctx context.Context, req *pb.UpdatePasswordRequest) (*emptypb.Empty, error) {
	log := scope.Log(ctx).WithOperation(operationUpdatePassword)

	err := s.users.UpdatePassword(ctx, scope.UserActor(ctx), req.OldPassword, req.NewPassword)
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
	case errors.Is(err, errx.ErrorUserInvalidSession):
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

const operationDeleteMyUser = "delete_my_user"

func (s *UserServer) DeleteMyUser(ctx context.Context, _ *pb.DeleteMyUserRequest) (*emptypb.Empty, error) {
	log := scope.Log(ctx).WithOperation(operationDeleteMyUser)

	err := s.users.DeleteMyUser(ctx, scope.UserActor(ctx))
	switch {
	case errors.Is(err, errx.ErrorUserInvalidSession):
		log.Warn("invalid session", "error", err)
		return nil, status.Error(codes.Unauthenticated, "invalid session")
	case err != nil:
		log.Error("unexpected error", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	default:
		log.Info("user deleted")
		return &emptypb.Empty{}, nil
	}
}
