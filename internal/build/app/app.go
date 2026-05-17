package app

import (
	"github.com/netbill/auth-svc/internal/build/config"
	"github.com/netbill/auth-svc/pkg/log"
)

type App struct {
	log    *log.Logger
	config *config.Config
}

func New(log *log.Logger, cfg *config.Config) *App {
	return &App{
		log:    log,
		config: cfg,
	}
}
