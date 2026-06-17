package mail

import (
	"fmt"
	"net/smtp"

	"github.com/jordan-wright/email"
)

type Config struct {
	Port     int
	Host     string
	Username string
	Password string
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
	email := email.NewEmail()
	email.From = m.config.Username
	email.To = to
	email.Subject = subject
	email.HTML = []byte(html)
	err := email.Send(
		fmt.Sprintf("%s:%d", m.config.Host, m.config.Port),
		smtp.PlainAuth(
			"",
			m.config.Username,
			m.config.Password,
			m.config.Host,
		),
	)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}
