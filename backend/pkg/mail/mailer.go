// Package mail provides transactional email delivery capabilities using SMTP.
package mail

import (
	"context"
	"crypto/tls"
	"fmt"
	netmail "net/mail"
	"net/smtp"
	"strings"
	"time"

	"github.com/jordan-wright/email"
)

// Config specifies connection credentials and sender information for SMTP operations.
type Config struct {
	// Port is the SMTP server port (e.g. 587 for STARTTLS, 465 for SSL, 25 for plain).
	Port int

	// Host is the SMTP server hostname.
	Host string

	// Username is the SMTP authentication username.
	Username string

	// Password is the SMTP authentication password.
	Password string

	// FromEmail is the default sender email address.
	FromEmail string

	// FromName is the human-readable display name for the sender.
	FromName string
}

// Mailer encapsulates outbound email transport using the specified Config.
type Mailer struct {
	config *Config
}

// NewMailer creates a new Mailer instance with the given configuration.
func NewMailer(config *Config) *Mailer {
	return &Mailer{
		config: config,
	}
}

// sanitizeHeader strips CRLF characters (\r and \n) to prevent header injection attacks.
func sanitizeHeader(s string) string {
	s = strings.ReplaceAll(s, "\r", "")
	return strings.ReplaceAll(s, "\n", "")
}

// SendHTML constructs and sends an HTML-formatted email message using a default 30-second timeout.
func (m *Mailer) SendHTML(
	to []string,
	subject, html string,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return m.SendHTMLWithContext(ctx, to, subject, html)
}

// SendHTMLWithContext constructs and sends an HTML-formatted email honoring context timeout and cancellation.
func (m *Mailer) SendHTMLWithContext(
	ctx context.Context,
	to []string,
	subject, html string,
) error {
	if m == nil || m.config == nil {
		return fmt.Errorf("mailer or mailer config is nil")
	}

	if len(to) == 0 {
		return fmt.Errorf("recipient list 'to' must not be empty")
	}

	for _, addr := range to {
		if _, err := netmail.ParseAddress(addr); err != nil {
			return fmt.Errorf("invalid recipient email address %q: %w", addr, err)
		}
	}

	from := m.config.FromEmail
	if from == "" {
		from = m.config.Username
	}
	if from == "" {
		from = "no-reply@prim.local"
	}
	from = sanitizeHeader(from)

	if m.config.FromName != "" {
		cleanName := sanitizeHeader(m.config.FromName)
		from = fmt.Sprintf("%s <%s>", cleanName, from)
	}

	e := email.NewEmail()
	e.From = from
	e.To = to
	e.Subject = sanitizeHeader(subject)
	e.HTML = []byte(html)

	var auth smtp.Auth
	if m.config.Username != "" && m.config.Password != "" {
		auth = smtp.PlainAuth(
			"",
			m.config.Username,
			m.config.Password,
			m.config.Host,
		)
	}

	addr := fmt.Sprintf("%s:%d", m.config.Host, m.config.Port)
	tlsConfig := &tls.Config{
		ServerName: m.config.Host,
		MinVersion: tls.VersionTLS12,
	}

	sendFn := func() error {
		switch m.config.Port {
		case 465:
			return e.SendWithTLS(addr, auth, tlsConfig)
		case 587:
			return e.SendWithStartTLS(addr, auth, tlsConfig)
		default:
			if auth != nil && m.config.Host != "localhost" && m.config.Host != "127.0.0.1" {
				return e.SendWithStartTLS(addr, auth, tlsConfig)
			}
			return e.Send(addr, auth)
		}
	}

	done := make(chan error, 1)
	go func() {
		done <- sendFn()
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("email sending canceled: %w", ctx.Err())
	case err := <-done:
		if err != nil {
			return fmt.Errorf("failed to send email: %w", err)
		}
		return nil
	}
}
