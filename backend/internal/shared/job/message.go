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

// Kafka Notification Topics
const (
	TopicNotificationsEmail = "notifications.email"
	TopicNotificationsInApp = "notifications.inapp"
	TopicNotificationsDLQ   = "notifications.dlq"
)

// NotificationEvent represents the canonical message payload sent through Kafka.
type NotificationEvent struct {
	EventID    uuid.UUID       `json:"eventId"`
	Type       string          `json:"type"`          // "email" or "in_app"
	Category   string          `json:"category"`      // "auth", "order", "system"
	UserID     *uuid.UUID      `json:"userId,omitempty"`
	Recipient  string          `json:"recipient"`     // Email address or User UUID string
	Locale     string          `json:"locale"`        // "ar" or "en"
	Template   string          `json:"template"`      // e.g. "email_otp", "welcome", "reset_password", "order_confirmed"
	Data       json.RawMessage `json:"data"`          // dynamic parameters
	Timestamp  time.Time       `json:"timestamp"`
	RetryCount int             `json:"retryCount"`
	LastError  string          `json:"lastError,omitempty"`
}
