package controller

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/api/grpc/reponses"
	"github.com/netbill/auth-svc/internal/api/grpc/scope"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/internal/modules/session"
	"github.com/netbill/auth-svc/pkg/pb"
	"github.com/netbill/restkit/pagi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type SessionCore interface {
	LoginByEmail(ctx context.Context, email, password string) (models.TokensPair, error)
	LoginByGoogle(ctx context.Context, email string) (models.TokensPair, error)
	LoginByUsername(ctx context.Context, username, password string) (models.TokensPair, error)
	Refresh(ctx context.Context, oldRefreshToken string) (models.TokensPair, error)
	GetMySession(ctx context.Context, actor models.AccountActor, sessionID uuid.UUID) (models.Session, error)
	GetMySessions(ctx context.Context, actor models.AccountActor, opts ...session.ListSessionsOption) (pagi.Page[[]models.Session], error)
	Logout(ctx context.Context, actor models.AccountActor) error
	DeleteMySession(ctx context.Context, actor models.AccountActor, sessionID uuid.UUID) error
	DeleteMySessions(ctx context.Context, actor models.AccountActor) error
}

type SessionMetrics interface {
	RecordEmailLogin(ctx context.Context, err *error)
	RecordUsernameLogin(ctx context.Context, err *error)
	RecordGoogleLogin(ctx context.Context, err *error)
	RecordTokenRefresh(ctx context.Context, err *error)
	RecordSessionDeleted(ctx context.Context, scope string, err *error)
}

// GoogleIDVerifier verifies a Google-issued OpenID Connect ID token and
// returns the verified email address it carries.
type GoogleIDVerifier interface {
	Verify(ctx context.Context, idToken string) (email string, err error)
}

type SessionServer struct {
	pb.UnimplementedSessionServiceServer
	sessions SessionCore
	metrics  SessionMetrics
	google   GoogleIDVerifier
}

func NewSessionServer(sessions SessionCore, m SessionMetrics, google GoogleIDVerifier) *SessionServer {
	return &SessionServer{sessions: sessions, metrics: m, google: google}
}

const operationLoginByEmail = "login_by_email"

func (s *SessionServer) LoginByEmail(ctx context.Context, req *pb.LoginByEmailRequest) (_ *pb.LoginResponse, err error) {
	log := scope.Log(ctx).WithOperation(operationLoginByEmail)
	defer s.metrics.RecordEmailLogin(ctx, &err)

	pair, err := s.sessions.LoginByEmail(ctx, req.Email, req.Password)
	switch {
	case errors.Is(err, errx.ErrorAccountNotFound),
		errors.Is(err, errx.ErrorAccountDeleted):
		log.Warn("account not found", "error", err)
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	case errors.Is(err, errx.ErrorPasswordInvalid):
		log.Warn("invalid password", "error", err)
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	case err != nil:
		log.Error("unexpected error", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	default:
		log.Info("login successful")
		return &pb.LoginResponse{Tokens: reponses.TokensPair(pair)}, nil
	}
}

const operationLoginByUsername = "login_by_username"

func (s *SessionServer) LoginByUsername(ctx context.Context, req *pb.LoginByUsernameRequest) (_ *pb.LoginResponse, err error) {
	log := scope.Log(ctx).WithOperation(operationLoginByUsername)
	defer s.metrics.RecordUsernameLogin(ctx, &err)

	pair, err := s.sessions.LoginByUsername(ctx, req.Username, req.Password)
	switch {
	case errors.Is(err, errx.ErrorAccountNotFound),
		errors.Is(err, errx.ErrorAccountDeleted):
		log.Warn("account not found", "error", err)
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	case errors.Is(err, errx.ErrorPasswordInvalid):
		log.Warn("invalid password", "error", err)
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	case err != nil:
		log.Error("unexpected error", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	default:
		log.Info("login successful")
		return &pb.LoginResponse{Tokens: reponses.TokensPair(pair)}, nil
	}
}

const operationLoginByGoogle = "login_by_google"

func (s *SessionServer) LoginByGoogle(ctx context.Context, req *pb.LoginByGoogleRequest) (_ *pb.LoginResponse, err error) {
	log := scope.Log(ctx).WithOperation(operationLoginByGoogle)
	defer s.metrics.RecordGoogleLogin(ctx, &err)

	email, err := s.google.Verify(ctx, req.IdToken)
	if err != nil {
		log.Warn("invalid google id token", "error", err)
		return nil, status.Error(codes.Unauthenticated, "invalid google id token")
	}

	pair, err := s.sessions.LoginByGoogle(ctx, email)
	switch {
	case errors.Is(err, errx.ErrorAccountNotFound),
		errors.Is(err, errx.ErrorAccountDeleted):
		log.Warn("account not found", "error", err)
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	case err != nil:
		log.Error("unexpected error", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	default:
		log.Info("login successful")
		return &pb.LoginResponse{Tokens: reponses.TokensPair(pair)}, nil
	}
}

const operationRefresh = "refresh"

func (s *SessionServer) Refresh(ctx context.Context, req *pb.RefreshRequest) (_ *pb.LoginResponse, err error) {
	log := scope.Log(ctx).WithOperation(operationRefresh)
	defer s.metrics.RecordTokenRefresh(ctx, &err)

	pair, err := s.sessions.Refresh(ctx, req.RefreshToken)
	switch {
	case errors.Is(err, errx.ErrorSessionExpired),
		errors.Is(err, errx.ErrorSessionTokenMismatch),
		errors.Is(err, errx.ErrorSessionNotFound),
		errors.Is(err, errx.ErrorAccountNotFound):
		log.Warn("invalid refresh token", "error", err)
		return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
	case err != nil:
		log.Error("unexpected error", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	default:
		log.Info("session refreshed")
		return &pb.LoginResponse{Tokens: reponses.TokensPair(pair)}, nil
	}
}

const operationGetMySession = "get_my_session"

func (s *SessionServer) GetMySession(ctx context.Context, req *pb.GetMySessionRequest) (*pb.GetMySessionResponse, error) {
	log := scope.Log(ctx).WithOperation(operationGetMySession)

	sessionID, err := uuid.Parse(req.SessionId)
	if err != nil {
		log.Warn("invalid session_id", "session_id", req.SessionId)
		return nil, status.Error(codes.InvalidArgument, "invalid session_id")
	}

	sess, err := s.sessions.GetMySession(ctx, scope.AccountActor(ctx), sessionID)
	switch {
	case errors.Is(err, errx.ErrorSessionNotFound),
		errors.Is(err, errx.ErrorSessionDeleted):
		log.Warn("session not found", "error", err)
		return nil, status.Error(codes.NotFound, "session not found")
	case err != nil:
		log.Error("unexpected error", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	default:
		log.Info("session retrieved")
		return &pb.GetMySessionResponse{Session: reponses.Session(sess)}, nil
	}
}

const operationGetMySessions = "get_my_sessions"

func (s *SessionServer) GetMySessions(ctx context.Context, req *pb.GetMySessionsRequest) (*pb.GetMySessionsResponse, error) {
	log := scope.Log(ctx).WithOperation(operationGetMySessions)

	var page, perPage uint32 = 1, 20
	if p := req.Pagination; p != nil {
		if p.Page > 0 {
			page = uint32(p.Page)
		}
		if p.PerPage > 0 {
			perPage = uint32(p.PerPage)
		}
	}

	opts := []session.ListSessionsOption{
		session.WithLimit(uint(perPage)),
		session.WithOffset(uint((page - 1) * perPage)),
		session.WithDeleted(pbToDeletedFilter(req.Filter)),
		session.WithLastUsedOrder(pbToLastUsedOrder(req.Order)),
	}

	result, err := s.sessions.GetMySessions(ctx, scope.AccountActor(ctx), opts...)
	switch {
	case err != nil:
		log.Error("unexpected error", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	default:
		log.Info("sessions retrieved")

		sessions := make([]*pb.Session, len(result.Data))
		for i, sess := range result.Data {
			sessions[i] = reponses.Session(sess)
		}

		return &pb.GetMySessionsResponse{
			Sessions: sessions,
			PageInfo: &pb.PageInfo{
				Total:      int32(result.Total),
				Page:       int32(page),
				PerPage:    int32(perPage),
				TotalPages: int32((result.Total + uint(perPage) - 1) / uint(perPage)),
			},
		}, nil
	}
}

const operationLogout = "logout"

func (s *SessionServer) Logout(ctx context.Context, _ *pb.LogoutRequest) (*emptypb.Empty, error) {
	log := scope.Log(ctx).WithOperation(operationLogout)

	err := s.sessions.Logout(ctx, scope.AccountActor(ctx))
	switch {
	case errors.Is(err, errx.ErrorSessionNotFound):
		log.Debug("session already deleted, treating as success")
		return &emptypb.Empty{}, nil
	case err != nil:
		log.Error("unexpected error", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	default:
		log.Info("logout successful")
		return &emptypb.Empty{}, nil
	}
}

const operationDeleteMySession = "delete_my_session"

func (s *SessionServer) DeleteMySession(ctx context.Context, req *pb.DeleteMySessionRequest) (*emptypb.Empty, error) {
	log := scope.Log(ctx).WithOperation(operationDeleteMySession)

	sessionID, err := uuid.Parse(req.SessionId)
	if err != nil {
		log.Warn("invalid session_id", "session_id", req.SessionId)
		return nil, status.Error(codes.InvalidArgument, "invalid session_id")
	}

	err = s.sessions.DeleteMySession(ctx, scope.AccountActor(ctx), sessionID)
	switch {
	case errors.Is(err, errx.ErrorSessionNotFound):
		log.Warn("session not found", "error", err)
		return nil, status.Error(codes.NotFound, "session not found")
	case errors.Is(err, errx.ErrorAccountInvalidSession):
		log.Warn("invalid session", "error", err)
		return nil, status.Error(codes.Unauthenticated, "invalid session")
	case err != nil:
		log.Error("unexpected error", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	default:
		log.Info("session deleted")
		return &emptypb.Empty{}, nil
	}
}

const operationDeleteMySessions = "delete_my_sessions"

func (s *SessionServer) DeleteMySessions(ctx context.Context, _ *pb.DeleteMySessionsRequest) (*emptypb.Empty, error) {
	log := scope.Log(ctx).WithOperation(operationDeleteMySessions)

	err := s.sessions.DeleteMySessions(ctx, scope.AccountActor(ctx))
	switch {
	case errors.Is(err, errx.ErrorAccountInvalidSession),
		errors.Is(err, errx.ErrorSessionNotFound):
		log.Warn("invalid session", "error", err)
		return nil, status.Error(codes.Unauthenticated, "invalid session")
	case err != nil:
		log.Error("unexpected error", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	default:
		log.Info("sessions deleted")
		return &emptypb.Empty{}, nil
	}
}

func pbToDeletedFilter(f pb.SessionDeletedFilter) session.DeletedFilter {
	switch f {
	case pb.SessionDeletedFilter_SESSION_DELETED_FILTER_ACTIVE:
		return session.DeletedFilterActive
	case pb.SessionDeletedFilter_SESSION_DELETED_FILTER_DELETED:
		return session.DeletedFilterDeleted
	default:
		return session.DeletedFilterAll
	}
}

func pbToLastUsedOrder(o pb.SessionLastUsedOrder) session.LastUsedOrder {
	if o == pb.SessionLastUsedOrder_SESSION_LAST_USED_ORDER_ASC {
		return session.LastUsedAsc
	}
	return session.LastUsedDesc
}
