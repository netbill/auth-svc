package cli

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/alecthomas/kingpin"
	"github.com/netbill/auth-svc/cli/maintenance"
	"github.com/netbill/auth-svc/cli/migrations"
	"github.com/netbill/logium"
	"github.com/sirupsen/logrus"
)

const ServiceName = "auth-svc"

func Run(args []string) bool {
	cfg, err := LoadConfig()
	if err != nil {
		logium.Fatalf("failed to load config: %v", err)
	}

	logium.SetLevel(logrus.DebugLevel)

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

	var (
		service = kingpin.New(ServiceName, "")

		runCmd     = service.Command("run", "run command")
		serviceCmd = runCmd.Command("service", "run service")

		migrateCmd     = service.Command("migrate", "migrate command")
		migrateUpCmd   = migrateCmd.Command("up", "migrate db up")
		migrateDownCmd = migrateCmd.Command("down", "migrate db down")

		eventsCmd = service.Command("events", "events commands")

		eventsInbox        = eventsCmd.Command("inbox", "inbox events")
		eventsInboxCleanup = eventsInbox.Command("cleanup", "cleanup inbox events")
		eventsInboxFailed  = eventsInboxCleanup.Command(
			"failed", "cleanup inbox events in failed state",
		)
		eventsInboxProcessing = eventsInboxCleanup.Command(
			"processing", "cleanup inbox events stuck in processing state",
		)
		eventsInboxProcessingProcessIDs = eventsInboxProcessing.Flag(
			"process-id",
			"cleanup only events reserved by this process id (repeatable)",
		).Strings()

		eventsOutbox        = eventsCmd.Command("outbox", "outbox events")
		eventsOutboxCleanup = eventsOutbox.Command("cleanup", "cleanup outbox events")
		eventsOutboxFailed  = eventsOutboxCleanup.Command(
			"failed", "cleanup outbox events in failed state",
		)
		eventsOutboxProcessing = eventsOutboxCleanup.Command(
			"processing", "cleanup outbox events stuck in processing state",
		)
		eventsOutboxProcessingProcessIDs = eventsOutboxProcessing.Flag(
			"process-id",
			"cleanup only events reserved by this process id (repeatable)",
		).Strings()
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup

	command, err := service.Parse(args[1:])
	if err != nil {
		log.WithError(err).Error("failed to parse arguments")
		return false
	}

	base := log.WithField("service", ServiceName)

	switch command {
	case serviceCmd.FullCommand():
		StartServices(ctx, cfg, base, &wg)
	case migrateUpCmd.FullCommand():
		err = migrations.MigrateUp(ctx, base, cfg.Database.SQL.URL)
	case migrateDownCmd.FullCommand():
		err = migrations.MigrateDown(ctx, base, cfg.Database.SQL.URL)
	case eventsOutboxFailed.FullCommand():
		err = maintenance.CleanupOutboxFailed(ctx, base, cfg.Database.SQL.URL)
	case eventsOutboxProcessing.FullCommand():
		err = maintenance.CleanupOutboxProcessing(ctx, base, cfg.Database.SQL.URL, *eventsOutboxProcessingProcessIDs...)
	case eventsInboxFailed.FullCommand():
		err = maintenance.CleanupInboxFailed(ctx, base, cfg.Database.SQL.URL)
	case eventsInboxProcessing.FullCommand():
		err = maintenance.CleanupInboxProcessing(ctx, base, cfg.Database.SQL.URL, *eventsInboxProcessingProcessIDs...)
	default:
		base.Errorf("unknown command %s", command)
		return false
	}
	if err != nil {
		base.WithError(err).Error("failed to exec cmd")
		return false
	}

	wgch := make(chan struct{})
	go func() {
		wg.Wait()
		close(wgch)
	}()

	select {
	case <-ctx.Done():
		base.Infof("Interrupt signal received: %v", ctx.Err())
		<-wgch
	case <-wgch:
		base.Info("All services stopped")
	}

	return true
}
