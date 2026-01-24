package cli

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/alecthomas/kingpin"
	"github.com/netbill/auth-svc/cmd"
	"github.com/netbill/auth-svc/cmd/migrations"
	"github.com/netbill/auth-svc/internal"
	"github.com/netbill/logium"
	"github.com/sirupsen/logrus"
)

func Run(args []string) bool {
	cfg, err := internal.LoadConfig()
	if err != nil {
		panic(err)
	}

	log := logium.New()

	lvl, err := logrus.ParseLevel(cfg.Log.Level)
	if err != nil {
		lvl = logrus.InfoLevel
		log.WithField("bad_level", cfg.Log.Level).Warn("unknown log level, fallback to info")
	}
	log.SetLevel(lvl)

	switch {
	case cfg.Log.Format == "json":
		log.SetFormatter(&logrus.JSONFormatter{})
	default:
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	log.Info("Starting server...")

	var (
		service        = kingpin.New("auth-svc", "")
		runCmd         = service.Command("run", "run command")
		serviceCmd     = runCmd.Command("service", "run service")
		migrateCmd     = service.Command("migrate", "migrate command")
		migrateUpCmd   = migrateCmd.Command("up", "migrate db up")
		migrateDownCmd = migrateCmd.Command("down", "migrate db down")
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	c, err := service.Parse(args[1:])
	if err != nil {
		log.WithError(err).Error("failed to parse arguments")
		return false
	}

	var wg sync.WaitGroup

	switch c {
	case serviceCmd.FullCommand():
		cmd.StartServices(ctx, cfg, log, &wg)
	case migrateUpCmd.FullCommand():
		err = migrations.MigrateUp(cfg.Database.SQL.URL)
	case migrateDownCmd.FullCommand():
		err = migrations.MigrateDown(cfg.Database.SQL.URL)
	default:
		log.Errorf("unknown command %s", c)
		return false
	}
	if err != nil {
		log.WithError(err).Error("failed to exec cmd")
		return false
	}

	wgch := make(chan struct{})
	go func() {
		wg.Wait()
		close(wgch)
	}()

	select {
	case <-ctx.Done():
		log.Warnf("Interrupt signal received: %v", ctx.Err())
		<-wgch
	case <-wgch:
		log.Warnf("All services stopped")
	}

	return true
}
