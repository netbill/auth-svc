package rest

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/netbill/logium"
	"github.com/netbill/restkit/tokens"
)

type Handlers interface {
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
	GetMyEmailData(w http.ResponseWriter, r *http.Request)

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
}

type Router struct {
	handlers    Handlers
	middlewares Middlewares
	log         *logium.Logger
}

func New(
	log *logium.Logger,
	middlewares Middlewares,
	handlers Handlers,
) *Router {
	return &Router{
		log:         log,
		middlewares: middlewares,
		handlers:    handlers,
	}
}

type Config struct {
	Port              string
	TimeoutRead       time.Duration
	TimeoutReadHeader time.Duration
	TimeoutWrite      time.Duration
	TimeoutIdle       time.Duration
}

func (rt *Router) Run(ctx context.Context, cfg Config) {
	auth := rt.middlewares.AccountAuth()
	sysadmin := rt.middlewares.AccountAuth(tokens.RoleSystemAdmin)

	r := chi.NewRouter()

	// CORS for swagger UI documentation need to delete after configuring nginx
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5001"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Route("/auth-svc", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {

			r.Route("/registration", func(r chi.Router) {
				r.Post("/", rt.handlers.Registration)
				r.With(auth, sysadmin).Post("/admin", rt.handlers.RegistrationByAdmin)
			})

			r.Route("/login", func(r chi.Router) {
				r.Post("/email", rt.handlers.LoginByEmail)
				r.Post("/username", rt.handlers.LoginByUsername)

				r.Route("/google", func(r chi.Router) {
					r.Post("/", rt.handlers.LoginByGoogleOAuth)
					r.Post("/callback", rt.handlers.LoginByGoogleOAuthCallback)
				})
			})

			r.Post("/refresh", rt.handlers.RefreshSession)

			r.With(auth).Route("/me", func(r chi.Router) {
				r.With(auth).Get("/", rt.handlers.GetMyAccount)
				r.With(auth).Delete("/", rt.handlers.DeleteMyAccount)

				r.With(auth).Get("/email", rt.handlers.GetMyEmailData)
				r.With(auth).Post("/logout", rt.handlers.Logout)
				r.With(auth).Post("/password", rt.handlers.UpdatePassword)
				r.With(auth).Post("/username", rt.handlers.UpdateUsername)

				r.With(auth).Route("/sessions", func(r chi.Router) {
					r.Get("/", rt.handlers.GetMySessions)
					r.Delete("/", rt.handlers.DeleteMySessions)

					r.Route("/{session_id}", func(r chi.Router) {
						r.Get("/", rt.handlers.GetMySession)
						r.Delete("/", rt.handlers.DeleteMySession)
					})
				})
			})
		})
	})

	srv := &http.Server{
		Addr:              cfg.Port,
		Handler:           r,
		ReadTimeout:       cfg.TimeoutRead,
		ReadHeaderTimeout: cfg.TimeoutReadHeader,
		WriteTimeout:      cfg.TimeoutWrite,
		IdleTimeout:       cfg.TimeoutIdle,
	}

	rt.log.Infof("starting REST service on %s", cfg.Port)

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
		rt.log.Warnf("shutting down REST service...")
	case err := <-errCh:
		if err != nil {
			rt.log.Errorf("REST server error: %v", err)
		}
	}

	shCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shCtx); err != nil {
		rt.log.Errorf("REST shutdown error: %v", err)
	} else {
		rt.log.Warnf("REST server stopped")
	}
}
