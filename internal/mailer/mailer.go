package mailer

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
)

type Config struct {
	Host   string
	Port   string
	User   string
	Pass   string
	From   string
	Secure bool
}

type Mailer struct {
	cfg Config
}

func New(cfg Config) *Mailer {
	return &Mailer{cfg: cfg}
}

func (m *Mailer) SendWelcomeEmail(to string) error {
	subject := "Welcome to Web Labs"
	body := "Hello!\r\n\r\nYour Web Labs account has been created successfully.\r\n"

	return m.send(to, subject, body)
}

func (m *Mailer) send(to, subject, body string) error {
	addr := net.JoinHostPort(m.cfg.Host, m.cfg.Port)
	auth := smtp.PlainAuth("", m.cfg.User, m.cfg.Pass, m.cfg.Host)
	message := buildMessage(m.cfg.From, to, subject, body)

	if m.cfg.Secure {
		return m.sendTLS(addr, auth, to, message)
	}

	return m.sendPlain(addr, auth, to, message)
}

func (m *Mailer) sendTLS(addr string, auth smtp.Auth, to string, message []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		ServerName: m.cfg.Host,
		MinVersion: tls.VersionTLS12,
	})
	if err != nil {
		return fmt.Errorf("connect SMTP over TLS: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, m.cfg.Host)
	if err != nil {
		return fmt.Errorf("create SMTP client: %w", err)
	}
	defer client.Quit()

	return m.sendWithClient(client, auth, to, message)
}

func (m *Mailer) sendPlain(addr string, auth smtp.Auth, to string, message []byte) error {
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("connect SMTP: %w", err)
	}
	defer client.Quit()

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{
			ServerName: m.cfg.Host,
			MinVersion: tls.VersionTLS12,
		}); err != nil {
			return fmt.Errorf("start SMTP TLS: %w", err)
		}
	}

	return m.sendWithClient(client, auth, to, message)
}

func (m *Mailer) sendWithClient(client *smtp.Client, auth smtp.Auth, to string, message []byte) error {
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("authenticate SMTP: %w", err)
	}
	if err := client.Mail(m.cfg.From); err != nil {
		return fmt.Errorf("set SMTP sender: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("set SMTP recipient: %w", err)
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("open SMTP data writer: %w", err)
	}

	if _, err := writer.Write(message); err != nil {
		_ = writer.Close()
		return fmt.Errorf("write SMTP message: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("close SMTP data writer: %w", err)
	}

	return nil
}

func buildMessage(from, to, subject, body string) []byte {
	headers := []string{
		"From: " + from,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		`Content-Type: text/plain; charset="UTF-8"`,
		"Content-Transfer-Encoding: 8bit",
	}

	return []byte(strings.Join(headers, "\r\n") + "\r\n\r\n" + body)
}
