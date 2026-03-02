package mailer

import (
	"context"
	"fmt"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	"github.com/mageas/the-punisher-backend/internal/platform/config"
)

// SMTPMailer sends transactional emails through a configured SMTP server.
type SMTPMailer struct {
	cfg config.SMTPConfig
}

func NewSMTPMailer(cfg config.SMTPConfig) *SMTPMailer {
	return &SMTPMailer{cfg: cfg}
}

func (m *SMTPMailer) SendConfirmationEmail(
	ctx context.Context,
	toEmail string,
	firstName string,
	confirmationURL string,
	expiresIn time.Duration,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	fromAddr := &mail.Address{Name: strings.TrimSpace(m.cfg.FromName), Address: strings.TrimSpace(m.cfg.FromEmail)}
	toAddr := &mail.Address{Address: strings.TrimSpace(toEmail)}

	if _, err := mail.ParseAddress(fromAddr.Address); err != nil {
		return fmt.Errorf("invalid smtp from email: %w", err)
	}
	if _, err := mail.ParseAddress(toAddr.Address); err != nil {
		return fmt.Errorf("invalid recipient email: %w", err)
	}

	serverAddr := fmt.Sprintf("%s:%d", m.cfg.Host, m.cfg.Port)
	subject := "Confirm your email address"
	body := fmt.Sprintf(
		"Hello %s,\r\n\r\nPlease confirm your email address by opening this link:\r\n%s\r\n\r\nThis link expires in %.0f hour(s).\r\n",
		strings.TrimSpace(firstName),
		confirmationURL,
		expiresIn.Hours(),
	)

	message := strings.Join([]string{
		fmt.Sprintf("From: %s", fromAddr.String()),
		fmt.Sprintf("To: %s", toAddr.String()),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	var auth smtp.Auth
	if strings.TrimSpace(m.cfg.Username) != "" {
		auth = smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)
	}

	if err := smtp.SendMail(serverAddr, auth, fromAddr.Address, []string{toAddr.Address}, []byte(message)); err != nil {
		return fmt.Errorf("failed to send confirmation email: %w", err)
	}

	return nil
}
