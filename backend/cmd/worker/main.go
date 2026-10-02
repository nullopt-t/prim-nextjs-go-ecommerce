package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/notifier"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/html"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/job"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/user"
	brokerkafka "github.com/m-mahmoud-alsaid/prim-backend/pkg/broker/kafka"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/log"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/mail"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()
	logger := log.NewConsoleLogger()

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

	txRunner := database.NewTxRunner(db)
	notificationRepo := user.NewNotificationRepository()
	notificationService := user.NewNotificationService(txRunner, notificationRepo, logger)

	// 2. Initialize Kafka Consumer & Producer
	kafkaCfg := brokerkafka.DefaultConfig(cfg.KafkaCfg.Brokers, cfg.KafkaCfg.GroupID)
	dlqProducer := brokerkafka.NewProducer(kafkaCfg)
	defer dlqProducer.Close()

	emailConsumer := brokerkafka.NewConsumer(kafkaCfg, job.TopicNotificationsEmail)
	defer emailConsumer.Close()

	inappConsumer := brokerkafka.NewConsumer(kafkaCfg, job.TopicNotificationsInApp)
	defer inappConsumer.Close()

	// 3. Initialize Mailer & HTML Template Renderer
	mailer := mail.NewMailer(&mail.Config{
		Host:     cfg.SMTPCfg.Host,
		Port:     cfg.SMTPCfg.Port,
		Username: cfg.SMTPCfg.Username,
		Password: cfg.SMTPCfg.Password,
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

	wg.Add(2)
	go emailWorker.Start(ctx, &wg)
	go inappWorker.Start(ctx, &wg)

	logger.Info("Kafka notification worker pools active and consuming messages...")

	<-ctx.Done()

	logger.Info("graceful shutdown initiated. Waiting for in-flight tasks to complete...")
	wg.Wait()
	logger.Info("all workers stopped cleanly. Exit.")
}
