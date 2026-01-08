package rest

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/netbill/auth-svc/internal"
	"github.com/netbill/logium"
	"github.com/netbill/restkit/roles"
)

type Handlers interface {
	Registration(w http.ResponseWriter, r *http.Request)
	RegistrationAdmin(w http.ResponseWriter, r *http.Request)

	LoginByEmail(w http.ResponseWriter, r *http.Request)
	LoginByUsername(w http.ResponseWriter, r *http.Request)
	LoginByGoogleOAuth(w http.ResponseWriter, r *http.Request)
	LoginByGoogleOAuthCallback(w http.ResponseWriter, r *http.Request)

	Logout(w http.ResponseWriter, r *http.Request)

	RefreshSession(w http.ResponseWriter, r *http.Request)

	GetMyAccount(w http.ResponseWriter, r *http.Request)
	GetMySession(w http.ResponseWriter, r *http.Request)
	GetMySessions(w http.ResponseWriter, r *http.Request)
	GetMyEmailData(w http.ResponseWriter, r *http.Request)

	UpdatePassword(w http.ResponseWriter, r *http.Request)
	UpdateUsername(w http.ResponseWriter, r *http.Request)

	DeleteMyAccount(w http.ResponseWriter, r *http.Request)
	DeleteMySession(w http.ResponseWriter, r *http.Request)
	DeleteMySessions(w http.ResponseWriter, r *http.Request)
}

type Middlewares interface {
	Auth() func(http.Handler) http.Handler
	RoleGrant(allowedRoles map[string]bool) func(http.Handler) http.Handler
}

type Service struct {
	handlers    Handlers
	middlewares Middlewares
	log         logium.Logger
}

func New(
	log logium.Logger,
	middlewares Middlewares,
	handlers Handlers,
) *Service {
	return &Service{
		log:         log,
		middlewares: middlewares,
		handlers:    handlers,
	}
}

func (s *Service) Run(ctx context.Context, cfg internal.Config) {
	auth := s.middlewares.Auth()
	sysadmin := s.middlewares.RoleGrant(map[string]bool{
		roles.SystemAdmin: true,
	})

	r := chi.NewRouter()

	r.Route("/auth-svc", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Post("/registration", s.handlers.Registration)

			r.Route("/login", func(r chi.Router) {
				r.Post("/email", s.handlers.LoginByEmail)
				r.Post("/username", s.handlers.LoginByUsername)

				r.Route("/google", func(r chi.Router) {
					r.Post("/", s.handlers.LoginByGoogleOAuth)
					r.Post("/callback", s.handlers.LoginByGoogleOAuthCallback)
				})
			})

			r.Post("/refresh", s.handlers.RefreshSession)

			r.With(auth).Route("/me", func(r chi.Router) {
				r.With(auth).Get("/", s.handlers.GetMyAccount)
				r.With(auth).Delete("/", s.handlers.DeleteMyAccount)

				r.With(auth).Get("/email", s.handlers.GetMyEmailData)
				r.With(auth).Post("/logout", s.handlers.Logout)
				r.With(auth).Post("/password", s.handlers.UpdatePassword)
				r.With(auth).Post("/username", s.handlers.UpdateUsername)

				r.With(auth).Route("/sessions", func(r chi.Router) {
					r.Get("/", s.handlers.GetMySessions)
					r.Delete("/", s.handlers.DeleteMySessions)

					r.Route("/{session_id}", func(r chi.Router) {
						r.Get("/", s.handlers.GetMySession)
						r.Delete("/", s.handlers.DeleteMySession)
					})
				})
			})

			r.Route("/admin", func(r chi.Router) {
				r.Use(auth)
				r.Use(sysadmin)

				r.Post("/", s.handlers.RegistrationAdmin)
			})
		})
	})

	srv := &http.Server{
		Addr:              cfg.Rest.Port,
		Handler:           r,
		ReadTimeout:       cfg.Rest.Timeouts.Read,
		ReadHeaderTimeout: cfg.Rest.Timeouts.ReadHeader,
		WriteTimeout:      cfg.Rest.Timeouts.Write,
		IdleTimeout:       cfg.Rest.Timeouts.Idle,
	}

	s.log.Infof("starting REST service on %s", cfg.Rest.Port)

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
		s.log.Warnf("shutting down REST service...")
	case err := <-errCh:
		if err != nil {
			s.log.Errorf("REST server error: %v", err)
		}
	}

	shCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shCtx); err != nil {
		s.log.Errorf("REST shutdown error: %v", err)
	} else {
		s.log.Warnf("REST server stopped")
	}
}
