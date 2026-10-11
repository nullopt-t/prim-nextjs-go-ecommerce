package mail

import (
	"fmt"
	"net/smtp"

	"github.com/jordan-wright/email"
)

type Config struct {
	Port      int
	Host      string
	Username  string
	Password  string
	FromEmail string
	FromName  string
}

type Mailer struct {
	config *Config
}

func NewMailer(config *Config) *Mailer {
	return &Mailer{
		config: config,
	}
}

func (m *Mailer) SendHTML(
	to []string,
	subject, html string,
) error {
	from := m.config.FromEmail
	if from == "" {
		from = m.config.Username
	}
	if from == "" {
		from = "no-reply@prim.local"
	}
	if m.config.FromName != "" {
		from = fmt.Sprintf("%s <%s>", m.config.FromName, from)
	}

	email := email.NewEmail()
	email.From = from
	email.To = to
	email.Subject = subject
	email.HTML = []byte(html)

	var auth smtp.Auth
	if m.config.Username != "" && m.config.Password != "" {
		auth = smtp.PlainAuth(
			"",
			m.config.Username,
			m.config.Password,
			m.config.Host,
		)
	}

	err := email.Send(
		fmt.Sprintf("%s:%d", m.config.Host, m.config.Port),
		auth,
	)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}
