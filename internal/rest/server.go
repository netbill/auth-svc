package rest

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/netbill/auth-svc/pkg/log"
	"github.com/netbill/restkit/tokens"
)

type Controller interface {
	Registration(w http.ResponseWriter, r *http.Request)
	RegistrationByAdmin(w http.ResponseWriter, r *http.Request)

	LoginByEmail(w http.ResponseWriter, r *http.Request)
	LoginByUsername(w http.ResponseWriter, r *http.Request)
	LoginByGoogleOAuth(w http.ResponseWriter, r *http.Request)
	LoginByGoogleOAuthCallback(w http.ResponseWriter, r *http.Request)

	Logout(w http.ResponseWriter, r *http.Request)

	RefreshSession(w http.ResponseWriter, r *http.Request)

	GetMyAccount(w http.ResponseWriter, r *http.Request)
	GetMySession(w http.ResponseWriter, r *http.Request)
	GetMySessions(w http.ResponseWriter, r *http.Request)

	UpdatePassword(w http.ResponseWriter, r *http.Request)
	UpdateUsername(w http.ResponseWriter, r *http.Request)

	DeleteMyAccount(w http.ResponseWriter, r *http.Request)
	DeleteMySession(w http.ResponseWriter, r *http.Request)
	DeleteMySessions(w http.ResponseWriter, r *http.Request)
}

type Middlewares interface {
	AccountAuth(
		allowedRoles ...string,
	) func(next http.Handler) http.Handler
	Logger(log *log.Logger) func(next http.Handler) http.Handler
	CorsDocs() func(next http.Handler) http.Handler
}

type Server struct {
	controller  Controller
	middlewares Middlewares
	log         *log.Logger
}

type ServerDeps struct {
	Controller  Controller
	Middlewares Middlewares
	Log         *log.Logger
}

func New(deps ServerDeps) *Server {
	return &Server{
		controller:  deps.Controller,
		middlewares: deps.Middlewares,
		log:         deps.Log,
	}
}

type Config struct {
	Port              int
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
}

func (s *Server) Run(ctx context.Context, cfg Config) {
	auth := s.middlewares.AccountAuth()
	sysadmin := s.middlewares.AccountAuth(tokens.RoleSystemAdmin)

	r := chi.NewRouter()
	r.Use(
		s.middlewares.Logger(s.log),
		s.middlewares.CorsDocs(),
	)

	r.Route("/auth-svc", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Route("/registration", func(r chi.Router) {
				r.Post("/", s.controller.Registration)
				r.With(auth, sysadmin).Post("/admin", s.controller.RegistrationByAdmin)
			})

			r.Route("/login", func(r chi.Router) {
				r.Post("/email", s.controller.LoginByEmail)
				r.Post("/username", s.controller.LoginByUsername)

				r.Route("/google", func(r chi.Router) {
					r.Post("/", s.controller.LoginByGoogleOAuth)
					r.Post("/callback", s.controller.LoginByGoogleOAuthCallback)
				})
			})

			r.Post("/refresh", s.controller.RefreshSession)

			r.With(auth).Route("/me", func(r chi.Router) {
				r.Get("/", s.controller.GetMyAccount)
				r.Delete("/", s.controller.DeleteMyAccount)

				r.Delete("/logout", s.controller.Logout)
				r.Patch("/password", s.controller.UpdatePassword)
				r.Patch("/username", s.controller.UpdateUsername)

				r.Route("/sessions", func(r chi.Router) {
					r.Get("/", s.controller.GetMySessions)
					r.Delete("/", s.controller.DeleteMySessions)

					r.Route("/{session_id}", func(r chi.Router) {
						r.Get("/", s.controller.GetMySession)
						r.Delete("/", s.controller.DeleteMySession)
					})
				})
			})
		})
	})

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           r,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	s.log.WithField("port", cfg.Port).Info("starting http service...")

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		} else {
			errCh <- nil
		}
	}()

	select {
	case <-ctx.Done():
		s.log.Info("shutting down http service...")
	case err := <-errCh:
		if err != nil {
			s.log.WithError(err).Error("http server error")
		}
	}

	shCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shCtx); err != nil {
		s.log.WithError(err).Error("failed to shutdown http server gracefully")
	} else {
		s.log.Info("http server shutdown gracefully")
	}
}
