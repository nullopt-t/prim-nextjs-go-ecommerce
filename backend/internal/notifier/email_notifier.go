package notifier

import (
	"context"
	"encoding/json"

	"github.com/nullopt-t/prim-backend/internal/shared/job"
	"github.com/nullopt-t/prim-backend/pkg/log"
)

type EmailNotifier struct {
	queue  job.JobQueue
	logger log.Logger
}

func NewEmailNotifier(
	queue job.JobQueue,
	logger log.Logger,
) *EmailNotifier {
	if queue == nil {
		panic("notifier: queue is required")
	}
	if logger == nil {
		panic("notifier: logger is required")
	}
	return &EmailNotifier{
		queue:  queue,
		logger: logger,
	}
}

func (n *EmailNotifier) NotifyOTP(
	ctx context.Context,
	channel,
	identifier,
	otp string,
) error {
	n.logger.Debug("notify otp",
		log.Meta{
			"identifier": identifier,
			"channel":    channel,
		},
	)

	payload, err := json.Marshal(
		OTPPayload{
			Identifier: identifier,
			Code:       otp,
		},
	)
	if err != nil {
		return err
	}

	var msgType job.MessageType
	switch channel {
	case "sms":
		msgType = job.MessageTypeSMS
	case "email":
		msgType = job.MessageTypeEmail
	}

	return n.queue.Enqueue(ctx,
		job.NewJobMessage(
			msgType,
			job.CommandEmailOTP,
			payload,
		),
	)
}

func (n *EmailNotifier) NotifyWelcome(
	ctx context.Context,
	email string,
) error {
	payload, err := json.Marshal(
		WelcomePayload{
			Email: email,
		},
	)
	if err != nil {
		return err
	}

	return n.queue.Enqueue(ctx,
		job.NewJobMessage(
			job.MessageTypeEmail,
			job.CommandWelcome,
			payload,
		),
	)
}

func (n *EmailNotifier) NotifyResetPassword(
	ctx context.Context,
	email, token string,
) error {
	payload, err := json.Marshal(
		ResetPasswordPayload{
			Email: email,
			Token: token,
		},
	)
	if err != nil {
		return err
	}
	return n.queue.Enqueue(ctx,
		job.NewJobMessage(
			job.MessageTypeEmail,
			job.CommandResetPassword,
			payload,
		),
	)
}
