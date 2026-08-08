package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/netbill/auth-svc/internal/api/grpc/controller"
	"github.com/netbill/auth-svc/internal/api/grpc/interceptors"
	"github.com/netbill/auth-svc/internal/modules/auth"
	"github.com/netbill/auth-svc/internal/observability/metrics"
	"github.com/netbill/auth-svc/pkg/log"
	"github.com/netbill/auth-svc/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	auth     *auth.Service
	accounts controller.AccountCore
	sessions controller.SessionCore
	google   controller.GoogleIDVerifier
	tokenMgr interceptors.TokenParser
	metrics  *metrics.Metrics
	log      *log.Logger
}

type ServerDeps struct {
	Auth     *auth.Service
	Accounts controller.AccountCore
	Sessions controller.SessionCore
	Google   controller.GoogleIDVerifier
	TokenMgr interceptors.TokenParser
	Metrics  *metrics.Metrics
	Log      *log.Logger
}

type Config struct {
	Port int
}

func New(deps ServerDeps) *Server {
	return &Server{
		auth:     deps.Auth,
		accounts: deps.Accounts,
		sessions: deps.Sessions,
		google:   deps.Google,
		metrics:  deps.Metrics,
		tokenMgr: deps.TokenMgr,
		log:      deps.Log,
	}
}

func (s *Server) Run(ctx context.Context, cfg Config) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		s.log.Error("grpc: failed to listen", "error", err)
		return
	}

	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.LogInterceptor(s.log),
			interceptors.AuthInterceptor(s.tokenMgr),
		),
	)
	pb.RegisterAuthServiceServer(srv, controller.NewAuthServer(s.auth, s.tokenMgr))
	pb.RegisterAccountServiceServer(srv, controller.NewAccountServer(s.accounts, s.metrics))
	pb.RegisterSessionServiceServer(srv, controller.NewSessionServer(s.sessions, s.metrics, s.google))
	reflection.Register(srv)

	s.log.Info("starting grpc server", "port", cfg.Port)

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Serve(lis)
	}()

	select {
	case <-ctx.Done():
		s.log.Info("shutting down grpc server...")
		srv.GracefulStop()
	case err := <-errCh:
		if err != nil {
			s.log.Error("grpc server error", "error", err)
		}
	}
}
