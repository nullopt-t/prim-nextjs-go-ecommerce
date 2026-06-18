package job

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type MessageType string

const (
	MessageTypeEmail MessageType = "email"
	MessageTypeSMS   MessageType = "sms"
)

type MessageStatus string

const (
	MessageStatusPending   MessageStatus = "pending"
	MessageStatusSuccess   MessageStatus = "success"
	MessageStatusFailure   MessageStatus = "failure"
	MessageStatusCancelled MessageStatus = "cancelled"
)

type Command string

const (
	CommandEmailOTP      Command = "email_otp"
	CommandWelcome       Command = "email_welcome"
	CommandResetPassword Command = "email_reset_password"
)

type JobMessage struct {
	ID         uuid.UUID       `json:"id"`
	Type       MessageType     `json:"type"`
	Command    Command         `json:"command"`
	Status     MessageStatus   `json:"status"`
	Attempts   int             `json:"attempts"`
	Payload    json.RawMessage `json:"payload"`
	EnqueuedAt time.Time       `json:"enqueued_at"`
}

func NewJobMessage(
	mtype MessageType,
	command Command,
	payload json.RawMessage,
) *JobMessage {
	return &JobMessage{
		ID:         uuid.New(),
		Type:       mtype,
		Command:    command,
		Status:     MessageStatusPending,
		Attempts:   0,
		Payload:    payload,
		EnqueuedAt: time.Now(),
	}
}
