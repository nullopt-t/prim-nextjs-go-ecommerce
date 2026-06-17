package main

import (
	"github.com/m-mahmoud-alsaid/prim-backend/internal/notifier"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/html"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/job"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/log"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/mail"
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()
	logger := log.NewConsoleLogger()

	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", cfg.RedisCfg.Host, cfg.RedisCfg.Port),
	})

	logger.Info("user client created")

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		logger.Error("redis ping failed", log.Meta{
			"Error": err,
		})
		return
	}
	logger.Info("redis ping successful")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// handle Ctrl+C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	go func() {
		<-sig
		cancel()
	}()

	jobQueue := job.NewJobQueue(redisClient, job.EmailQueue)

	mailer := mail.NewMailer(&mail.Config{
		Host:     cfg.SMTPCfg.Host,
		Port:     cfg.SMTPCfg.Port,
		Username: cfg.SMTPCfg.Username,
		Password: cfg.SMTPCfg.Password,
	})

	renderer, err := html.NewRenderer()
	if err != nil {
		logger.Error(
			"renderer creation failed:", log.Meta{
				"Error": err,
			})
		return
	}

	commandDispatcher := job.NewMessageDispatcher()
	emailMessageHandler := notifier.NewEmailHandler(renderer, commandDispatcher, mailer)
	commandDispatcher.Register(string(job.CommandEmailOTP), emailMessageHandler.SendEmailOTP)

	typeDispatcher := job.NewMessageDispatcher()
	typeDispatcher.Register(string(job.MessageTypeEmail), emailMessageHandler.HandleMessage)

	worker := job.NewNotifierWorker(
		jobQueue,
		typeDispatcher,
		logger,
		cfg.ClientCfg,
	)

	worker.Start(ctx)
}
