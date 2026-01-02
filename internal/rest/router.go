package rest

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/netbill/auth-svc/internal"
	"github.com/netbill/auth-svc/internal/rest/meta"
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
	Auth(userCtxKey interface{}, skUser string) func(http.Handler) http.Handler
	RoleGrant(userCtxKey interface{}, allowedRoles map[string]bool) func(http.Handler) http.Handler
}

func Run(ctx context.Context, cfg internal.Config, log logium.Logger, m Middlewares, h Handlers) {
	auth := m.Auth(meta.AccountDataCtxKey, cfg.JWT.User.AccessToken.SecretKey)
	sysadmin := m.RoleGrant(meta.AccountDataCtxKey, map[string]bool{
		roles.SystemAdmin: true,
	})

	r := chi.NewRouter()

	r.Route("/auth-svc", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Post("/registration", h.Registration)

			r.Route("/login", func(r chi.Router) {
				r.Post("/email", h.LoginByEmail)
				r.Post("/username", h.LoginByUsername)

				r.Route("/google", func(r chi.Router) {
					r.Post("/", h.LoginByGoogleOAuth)
					r.Post("/callback", h.LoginByGoogleOAuthCallback)
				})
			})

			r.Post("/refresh", h.RefreshSession)

			r.With(auth).Route("/me", func(r chi.Router) {
				r.With(auth).Get("/", h.GetMyAccount)
				r.With(auth).Delete("/", h.DeleteMyAccount)

				r.With(auth).Get("/email", h.GetMyEmailData)
				r.With(auth).Post("/logout", h.Logout)
				r.With(auth).Post("/password", h.UpdatePassword)
				r.With(auth).Post("/username", h.UpdateUsername)

				r.With(auth).Route("/sessions", func(r chi.Router) {
					r.Get("/", h.GetMySessions)
					r.Delete("/", h.DeleteMySessions)

					r.Route("/{session_id}", func(r chi.Router) {
						r.Get("/", h.GetMySession)
						r.Delete("/", h.DeleteMySession)
					})
				})
			})

			r.Route("/admin", func(r chi.Router) {
				r.Use(auth)
				r.Use(sysadmin)

				r.Post("/", h.RegistrationAdmin)
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

	log.Infof("starting REST service on %s", cfg.Rest.Port)

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
		log.Info("shutting down REST service...")
	case err := <-errCh:
		if err != nil {
			log.Errorf("REST server error: %v", err)
		}
	}

	shCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shCtx); err != nil {
		log.Errorf("REST shutdown error: %v", err)
	} else {
		log.Info("REST server stopped")
	}
}
