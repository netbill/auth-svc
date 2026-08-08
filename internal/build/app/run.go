package app

import (
	"context"
	"fmt"
	"sync"

	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	grpcapi "github.com/netbill/auth-svc/internal/api/grpc"
	"github.com/netbill/auth-svc/internal/api/rest"
	"github.com/netbill/auth-svc/internal/api/rest/controller"
	"github.com/netbill/auth-svc/internal/api/rest/middlewares"
	"github.com/netbill/auth-svc/internal/bus"
	"github.com/netbill/auth-svc/internal/media"
	authmodule "github.com/netbill/auth-svc/internal/modules/auth"
	"github.com/netbill/auth-svc/internal/modules/session"
	"github.com/netbill/auth-svc/internal/modules/user"
	"github.com/netbill/auth-svc/internal/observability/metrics"
	"github.com/netbill/auth-svc/internal/observability/telemetry"
	"github.com/netbill/auth-svc/internal/repo/chache"
	"github.com/netbill/auth-svc/internal/repo/pg"
	"github.com/netbill/auth-svc/pkg/googleid"
	"github.com/netbill/auth-svc/pkg/passmanager"
	"github.com/netbill/auth-svc/pkg/tokenmanager"
	"github.com/netbill/auth-svc/pkg/username"
	"github.com/netbill/awsx"
	"github.com/netbill/pgdbx"
)

func (a *App) Run(ctx context.Context) error {
	var wg sync.WaitGroup

	run := func(f func()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			f()
		}()
	}

	shutdownTelemetry, err := telemetry.Init(ctx, telemetry.Config{
		ServiceName:       a.config.Service.Name,
		CollectorEndpoint: a.config.OTEL.CollectorEndpoint,
		SamplingRatio:     a.config.OTEL.SamplingRatio,
	})
	if err != nil {
		return fmt.Errorf("init telemetry: %w", err)
	}
	defer shutdownTelemetry(ctx)

	svcMetrics, err := metrics.New()
	if err != nil {
		return fmt.Errorf("init metrics: %w", err)
	}

	pool, err := a.config.PoolDB(ctx)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer pool.Close()

	db := pgdbx.NewDB(pool)
	redisClient := a.config.RedisClient()
	defer redisClient.Close()

	userRepo := pg.NewUserRepo(db)
	emailRepo := pg.NewEmailRepo(db)
	passwordRepo := pg.NewPasswordRepo(db)
	sessionRepo := pg.NewSessionRepo(db)
	outboxRepo := pg.NewOutboxRepo(db, a.config.Kafka.Identity)

	redisTTL := a.config.Database.Redis.TTL

	userCache := chache.NewUserCache(redisClient, redisTTL.User, svcMetrics, a.log)
	emailCache := chache.NewEmailCache(redisClient, redisTTL.Email, svcMetrics, a.log)
	passwordCache := chache.NewPasswordCache(redisClient, redisTTL.Password, svcMetrics, a.log)
	sessionCache := chache.NewSessionCache(redisClient, redisTTL.Session, svcMetrics, a.log)
	qrCache := chache.NewQRCache(redisClient)

	qrPublisher := bus.NewPublisher(redisClient)
	qrSubscriber := bus.NewSubscriber(redisClient)

	passMgr := passmanager.New(a.config.Auth.PassBcryptCost)

	awsCfg, err := awscfg.LoadDefaultConfig(
		ctx,
		awscfg.WithRegion(a.config.S3.Aws.Region),
		awscfg.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				a.config.S3.Aws.AccessKeyID,
				a.config.S3.Aws.SecretAccessKey,
				a.config.S3.Aws.SessionToken,
			),
		),
	)
	if err != nil {
		return fmt.Errorf("load aws config: %w", err)
	}

	mediaStorage := media.NewStorage(awsx.New(a.config.S3.Aws.BucketName, awsCfg), media.Config{
		LinkTTL:    a.config.S3.Media.Link.TTL,
		UserAvatar: a.config.S3.Media.Resources.User.Avatar,
	})
	mediaResolver := media.NewResolver(a.config.S3.Aws.BaseURL)
	usernameValidator := username.NewValidator()

	tokenMgr := tokenmanager.New(tokenmanager.Config{
		Issuer:           a.config.Auth.Tokens.Issuer,
		AccessSecretKey:  a.config.Auth.Tokens.UserAccess.SecretKey,
		AccessTTL:        a.config.Auth.Tokens.UserAccess.TTL,
		RefreshSecretKey: a.config.Auth.Tokens.UserRefresh.SecretKey,
		RefreshTTL:       a.config.Auth.Tokens.UserRefresh.TTL,
		RefreshHashKey:   a.config.Auth.Tokens.UserRefresh.HashKey,
	})

	authSvc := authmodule.New(authmodule.ServiceDeps{
		UserRepo:    userRepo,
		SessionRepo: sessionRepo,
	})

	userSvc := user.New(user.ServiceDeps{
		Auth:          authSvc,
		UserRepo:      userRepo,
		EmailRepo:     emailRepo,
		PasswordRepo:  passwordRepo,
		SessionRepo:   sessionRepo,
		Tx:            db,
		UserCache:     userCache,
		EmailCache:    emailCache,
		PasswordCache: passwordCache,
		SessionsCache: sessionCache,
		PassManager:   passMgr,
		Messenger:     outboxRepo,
		Bucket:        mediaStorage,
		Username:      usernameValidator,
	})

	broker := bus.NewBroker(qrPublisher, qrSubscriber)

	sessionSvc := session.New(session.ServiceDeps{
		Auth:          authSvc,
		UserRepo:      userRepo,
		EmailRepo:     emailRepo,
		PasswordRepo:  passwordRepo,
		SessionRepo:   sessionRepo,
		Tx:            db,
		PasswordCache: passwordCache,
		UserCache:     userCache,
		SessionsCache: sessionCache,
		PassManager:   passMgr,
		TokenManager:  tokenMgr,
		QRStore:       qrCache,
		Bus:           broker,
	})

	userCtrl := controller.NewUserController(userSvc, svcMetrics)
	sessionCtrl := controller.NewSessionController(sessionSvc, a.config.GoogleOAuth(), svcMetrics, broker)

	mdll := middlewares.New(tokenMgr)
	router := rest.New(rest.ServerDeps{
		Users:       userCtrl,
		Sessions:    sessionCtrl,
		QR:          sessionCtrl,
		Middlewares: mdll,
		Log:         a.log,
		Resolver:    mediaResolver,
	})

	run(func() {
		router.Run(ctx, rest.Config{
			Port:              a.config.Rest.Port,
			ReadTimeout:       a.config.Rest.Timeouts.Read,
			ReadHeaderTimeout: a.config.Rest.Timeouts.ReadHeader,
			WriteTimeout:      a.config.Rest.Timeouts.Write,
			IdleTimeout:       a.config.Rest.Timeouts.Idle,
		})
	})

	googleVerifier := googleid.New(a.config.Auth.OAuth.Google.ClientID)

	grpcServer := grpcapi.New(grpcapi.ServerDeps{
		Auth:     authSvc,
		Users:    userSvc,
		Sessions: sessionSvc,
		Google:   googleVerifier,
		Metrics:  svcMetrics,
		TokenMgr: tokenMgr,
		Log:      a.log,
	})

	run(func() {
		grpcServer.Run(ctx, grpcapi.Config{
			Port: a.config.GRPC.Port,
		})
	})

	a.log.Info("starting application")
	wg.Wait()
	return nil
}
