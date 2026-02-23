package cli

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kingpin"
	"github.com/netbill/auth-svc/internal/app"
	"github.com/netbill/auth-svc/internal/config"
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
			"process-id", "cleanup only events reserved by this process id (repeatable)",
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
			"process-id", "cleanup only events reserved by this process id (repeatable)",
		).Strings()
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
	case eventsOutboxFailed.FullCommand():
		err = application.CleanupOutboxFailedEvents(ctx)
	case eventsOutboxProcessing.FullCommand():
		err = application.CleanupOutboxProcessingEvents(ctx, *eventsOutboxProcessingProcessIDs...)
	case eventsInboxFailed.FullCommand():
		err = application.CleanupInboxFailedEvents(ctx)
	case eventsInboxProcessing.FullCommand():
		err = application.CleanupInboxProcessingEvents(ctx, *eventsInboxProcessingProcessIDs...)
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
