package notifier

import (
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/job"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/log"
	"context"
	"encoding/json"
)

type EmailNotifier struct {
	queue  job.JobQueue
	logger log.Logger
}

func NewEmailNotifier(
	queue job.JobQueue,
	logger log.Logger,
) *EmailNotifier {
	return &EmailNotifier{
		queue:  queue,
		logger: logger,
	}
}

func (n *EmailNotifier) NotifyOTP(
	ctx context.Context,
	email, otp string,
) error {
	n.logger.Debug("notify otp",
		log.Meta{
			"email": email,
			"otp":   otp,
		},
	)

	payload, err := json.Marshal(
		OTPPayload{
			Email: email,
			Code:  otp,
		},
	)
	if err != nil {
		return err
	}

	return n.queue.Enqueue(ctx,
		job.NewJobMessage(
			job.MessageTypeEmail,
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
