package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/nullopt-t/prim-backend/internal/catalog/inventory"
	"github.com/nullopt-t/prim-backend/internal/notifier"
	"github.com/nullopt-t/prim-backend/internal/order"
	"github.com/nullopt-t/prim-backend/internal/setting"
	"github.com/nullopt-t/prim-backend/internal/shared/html"
	"github.com/nullopt-t/prim-backend/internal/shared/job"
	"github.com/nullopt-t/prim-backend/internal/user"
	brokerkafka "github.com/nullopt-t/prim-backend/pkg/broker/kafka"
	"github.com/nullopt-t/prim-backend/pkg/config"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/nullopt-t/prim-backend/pkg/log"
	"github.com/nullopt-t/prim-backend/pkg/mail"
	"github.com/nullopt-t/prim-backend/pkg/redisclient"
)

func main() {
	_ = config.LoadDotEnv(".env", "backend/.env", "../.env")
	logger := log.NewConsoleLogger()
	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load configuration", log.Meta{"error": err.Error()})
		return
	}

	logger.Info("starting Prim Notification Worker daemon...")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// 1. Connect to PostgreSQL for in-app alert persistence
	db, err := database.ConnectDB(ctx, cfg.DBCfg)
	if err != nil {
		logger.Error("failed to connect to PostgreSQL", log.Meta{"error": err})
		return
	}
	defer db.Close()

	// Redis is also the fallback transport for email notifications when Kafka
	// is unavailable. The client is lazy, so a temporary Redis outage does not
	// prevent the worker from starting; the fallback consumer will retry.
	redisClient, err := redisclient.New(cfg.RedisCfg)
	if err != nil {
		logger.Error("failed to configure redis client", log.Meta{"error": err})
		return
	}
	defer func() { _ = redisClient.Close() }()
	if err := redisClient.Ping(ctx); err != nil {
		logger.Warn("redis connection issue", log.Meta{"error": err})
	}

	txRunner := database.NewTxRunner(db)
	notificationRepo := user.NewNotificationRepository()
	notificationService := user.NewNotificationService(txRunner, notificationRepo, logger)

	// Reservation lifecycle: the reaper cancels pending orders whose checkout
	// hold expired and returns their held stock to salable. The complementary
	// branch (committing holds into sales on payment capture) runs in the API
	// process where payments are updated.
	inventoryRepo := inventory.NewRepository()
	// The worker only drives reservation expiry, which never resolves image
	// URLs, so the object URL resolver is intentionally nil here.
	inventoryService := inventory.NewService(logger, txRunner, inventoryRepo, nil)
	orderRepo := order.NewRepository()
	userRepo := user.NewPostgresRepository()

	// 2. Initialize Kafka Consumer & Producer
	kafkaCfg := brokerkafka.DefaultConfig(cfg.KafkaCfg.Brokers, cfg.KafkaCfg.GroupID)
	dlqProducer := brokerkafka.NewProducer(kafkaCfg)
	defer func() { _ = dlqProducer.Close() }()

	emailFallback := job.NewRedisEmailFallbackQueue(redisClient)
	inAppFallback := job.NewRedisInAppFallbackQueue(redisClient)
	orderNotifier := notifier.NewKafkaNotifier(dlqProducer, logger, emailFallback, inAppFallback)
	orderService := order.NewService(
		txRunner,
		orderRepo,
		logger,
		inventoryRepo,
		inventoryService,
		userRepo,
		setting.NewService(txRunner, setting.NewRepository()),
		orderNotifier,
	)
	reservationReaper := order.NewReservationReaper(orderService, logger, time.Minute)

	emailConsumer := brokerkafka.NewConsumer(kafkaCfg, job.TopicNotificationsEmail)
	defer func() { _ = emailConsumer.Close() }()

	inappConsumer := brokerkafka.NewConsumer(kafkaCfg, job.TopicNotificationsInApp)
	defer func() { _ = inappConsumer.Close() }()

	// 3. Initialize Mailer & HTML Template Renderer
	mailer := mail.NewMailer(&mail.Config{
		Host:      cfg.SMTPCfg.Host,
		Port:      cfg.SMTPCfg.Port,
		Username:  cfg.SMTPCfg.Username,
		Password:  cfg.SMTPCfg.Password,
		FromEmail: cfg.SMTPCfg.FromEmail,
		FromName:  cfg.SMTPCfg.FromName,
	})

	renderer, err := html.NewRenderer()
	if err != nil {
		logger.Error("failed to initialize html template renderer", log.Meta{"error": err})
		return
	}

	eventHandler := notifier.NewEventHandler(
		renderer,
		mailer,
		notificationService,
		logger,
		cfg.ClientCfg.BaseURL,
	)

	// 4. Start concurrent Kafka workers with graceful shutdown
	var wg sync.WaitGroup

	emailWorker := job.NewKafkaWorker(
		emailConsumer,
		dlqProducer,
		eventHandler.HandleEmailEvent,
		logger,
		job.TopicNotificationsEmail,
		3,
	)

	inappWorker := job.NewKafkaWorker(
		inappConsumer,
		dlqProducer,
		eventHandler.HandleInAppEvent,
		logger,
		job.TopicNotificationsInApp,
		3,
	)

	// Email events that the API could not publish to Kafka are consumed here.
	redisEmailWorker := job.NewRedisEmailFallbackWorker(
		redisClient,
		eventHandler.HandleEmailEvent,
		logger,
		2,
	)

	// In-app events that the API could not publish to Kafka are consumed here.
	redisInAppWorker := job.NewRedisInAppFallbackWorker(
		redisClient,
		eventHandler.HandleInAppEvent,
		logger,
		2,
	)

	wg.Add(5)
	go emailWorker.Start(ctx, &wg)
	go inappWorker.Start(ctx, &wg)
	go redisEmailWorker.Start(ctx, &wg)
	go redisInAppWorker.Start(ctx, &wg)
	go reservationReaper.Start(ctx, &wg)

	logger.Info("notification worker pools active (Kafka + Redis email/in-app fallbacks)...")

	<-ctx.Done()

	logger.Info("graceful shutdown initiated. Waiting for in-flight tasks to complete...")
	wg.Wait()
	logger.Info("all workers stopped cleanly. Exit.")
}
