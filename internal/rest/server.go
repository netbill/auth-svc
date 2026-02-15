package rest

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
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
	Logger(log *logium.Entry) func(next http.Handler) http.Handler
	CorsDocs() func(next http.Handler) http.Handler
}

type Server struct {
	handlers    Handlers
	middlewares Middlewares
}

func New(
	middlewares Middlewares,
	handlers Handlers,
) *Server {
	return &Server{
		middlewares: middlewares,
		handlers:    handlers,
	}
}

type Config struct {
	Port     string `mapstructure:"port"`
	Timeouts struct {
		Read       time.Duration `mapstructure:"read"`
		ReadHeader time.Duration `mapstructure:"read_header"`
		Write      time.Duration `mapstructure:"write"`
		Idle       time.Duration `mapstructure:"idle"`
	} `mapstructure:"timeouts"`
}

func (s *Server) Run(ctx context.Context, log *logium.Entry, cfg Config) {
	auth := s.middlewares.AccountAuth()
	sysadmin := s.middlewares.AccountAuth(tokens.RoleSystemAdmin)

	log = log.WithField("component", "rest")

	r := chi.NewRouter()
	r.Use(
		s.middlewares.Logger(log),
		s.middlewares.CorsDocs(),
	)

	r.Route("/auth-svc", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Route("/registration", func(r chi.Router) {
				r.Post("/", s.handlers.Registration)
				r.With(auth, sysadmin).Post("/admin", s.handlers.RegistrationByAdmin)
			})

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
				r.Get("/", s.handlers.GetMyAccount)
				r.Delete("/", s.handlers.DeleteMyAccount)

				r.Get("/email", s.handlers.GetMyEmailData)
				r.Post("/logout", s.handlers.Logout)
				r.Post("/password", s.handlers.UpdatePassword)
				r.Post("/username", s.handlers.UpdateUsername)

				r.Route("/sessions", func(r chi.Router) {
					r.Get("/", s.handlers.GetMySessions)
					r.Delete("/", s.handlers.DeleteMySessions)

					r.Route("/{session_id}", func(r chi.Router) {
						r.Get("/", s.handlers.GetMySession)
						r.Delete("/", s.handlers.DeleteMySession)
					})
				})
			})
		})
	})

	srv := &http.Server{
		Addr:              cfg.Port,
		Handler:           r,
		ReadTimeout:       cfg.Timeouts.Read,
		ReadHeaderTimeout: cfg.Timeouts.ReadHeader,
		WriteTimeout:      cfg.Timeouts.Write,
		IdleTimeout:       cfg.Timeouts.Idle,
	}

	log.Infof("starting http service on %s", cfg.Port)

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
		log.Infof("shutting down http service...")
	case err := <-errCh:
		if err != nil {
			log.Errorf("http server error: %v", err)
		}
	}

	shCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shCtx); err != nil {
		log.Errorf("http shutdown error: %v", err)
	} else {
		log.Infof("http server stopped")
	}
}
