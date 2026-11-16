package sms

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/nullopt-t/prim-backend/pkg/log"
)

// Sender abstracts outbound SMS delivery.
type Sender interface {
	Send(ctx context.Context, to string, message string) error
}

// Config specifies SMS provider settings.
type Config struct {
	Provider         string
	TwilioAccountSID string
	TwilioAuthToken  string
	TwilioFromPhone  string
}

// NewSender constructs a Sender implementation based on Config.
// If provider is "twilio" and credentials are valid, TwilioSender is returned.
// Otherwise, it falls back to LogSender.
func NewSender(cfg Config, logger log.Logger) Sender {
	if strings.ToLower(cfg.Provider) == "twilio" &&
		cfg.TwilioAccountSID != "" &&
		cfg.TwilioAuthToken != "" &&
		cfg.TwilioFromPhone != "" {
		return NewTwilioSender(cfg.TwilioAccountSID, cfg.TwilioAuthToken, cfg.TwilioFromPhone, logger)
	}
	return NewLogSender(logger)
}

// LogSender is a development/fallback sender that writes SMS to the application logger.
type LogSender struct {
	logger log.Logger
}

func NewLogSender(logger log.Logger) *LogSender {
	return &LogSender{logger: logger}
}

func (s *LogSender) Send(ctx context.Context, to, message string) error {
	s.logger.Info("sms dispatched (log driver)", log.Meta{
		"recipient": to,
		"message":   message,
	})
	return nil
}

// TwilioSender sends SMS messages via the Twilio REST API.
type TwilioSender struct {
	accountSID string
	authToken  string
	fromPhone  string
	httpClient *http.Client
	baseURL    string
	logger     log.Logger
}

// TwilioOption allows customizing the TwilioSender.
type TwilioOption func(*TwilioSender)

// WithHTTPClient overrides the default HTTP client (e.g. for testing).
func WithHTTPClient(client *http.Client) TwilioOption {
	return func(s *TwilioSender) {
		if client != nil {
			s.httpClient = client
		}
	}
}

// WithBaseURL overrides the Twilio API endpoint base URL (e.g. for mock servers).
func WithBaseURL(url string) TwilioOption {
	return func(s *TwilioSender) {
		if url != "" {
			s.baseURL = strings.TrimSuffix(url, "/")
		}
	}
}

func NewTwilioSender(accountSID, authToken, fromPhone string, logger log.Logger, opts ...TwilioOption) *TwilioSender {
	s := &TwilioSender{
		accountSID: accountSID,
		authToken:  authToken,
		fromPhone:  fromPhone,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    "https://api.twilio.com",
		logger:     logger,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *TwilioSender) Send(ctx context.Context, to, message string) error {
	if to == "" {
		return errors.New("sms: recipient phone number cannot be empty")
	}
	if message == "" {
		return errors.New("sms: message content cannot be empty")
	}

	endpoint := fmt.Sprintf("%s/2010-04-01/Accounts/%s/Messages.json", s.baseURL, s.accountSID)

	data := url.Values{}
	data.Set("To", to)
	data.Set("From", s.fromPhone)
	data.Set("Body", message)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("sms: create request: %w", err)
	}

	req.SetBasicAuth(s.accountSID, s.authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sms: http post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("sms: twilio api returned status %d: %s", resp.StatusCode, string(respBody))
	}

	s.logger.Info("sms delivered via twilio", log.Meta{
		"recipient": to,
	})
	return nil
}
