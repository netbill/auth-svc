package cli

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kingpin"
	"github.com/netbill/auth-svc/internal/build/app"
	"github.com/netbill/auth-svc/internal/build/config"
)

func Run(args []string) {
	cfg := config.LoadConfig()
	log := cfg.Logger()

	var (
		service    = kingpin.New(cfg.Service.Name, "")
		runCmd     = service.Command("run", "run command flags: service")
		serviceCmd = runCmd.Command("service", "starting all service processes")

		migrateCmd     = service.Command("migrate", "migrate command")
		migrateUpCmd   = migrateCmd.Command("up", "migrate db up")
		migrateDownCmd = migrateCmd.Command("down", "migrate db down")
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	command, err := service.Parse(args[1:])
	if err != nil {
		log.WithError(err).Error("failed to parse arguments")
		return
	}

	application := app.New(log, cfg)
	switch command {
	case serviceCmd.FullCommand():
		err = application.Run(ctx)
	case migrateUpCmd.FullCommand():
		err = application.MigrateUp(ctx)
	case migrateDownCmd.FullCommand():
		err = application.MigrateDown(ctx)
	default:
		log.Error("unknown command %s", command)
		return
	}
	if err != nil {
		log.WithError(err).Error("failed to exec cmd")
		return
	}

	log.Info("all processes finished successfully")
}
