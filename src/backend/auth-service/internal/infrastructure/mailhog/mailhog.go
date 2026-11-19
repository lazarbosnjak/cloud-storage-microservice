package mailhog

import (
	"fmt"
	"log/slog"
	"net/smtp"

	"auth-service/internal/startup/config"
)

type Server struct {
	config config.Config
}

func NewServer(config config.Config) *Server {
	return &Server{
		config: config,
	}
}

func (s *Server) SendVerifyRegistrationMail(to []string, verificationLink string) {
	subject := "[POUCH] Verify registration"

	// TODO: Add HTML template to body
	body := verificationLink

	s.sendMail(subject, body, to)
}

func (s *Server) sendMail(subject, body string, to []string) {
	username := s.config.SMTP.Username
	password := s.config.SMTP.Password
	host := s.config.SMTP.Host
	port := s.config.SMTP.Port

	from := username + "@pouch.com"

	// not needed for unencrypted
	_ = smtp.PlainAuth(
		"",
		username,
		password,
		host,
	)

	message := fmt.Sprintf("From: %s\r\n", username)
	message += fmt.Sprintf("To: %s\r\n", to)
	message += fmt.Sprintf("Subject: %s\r\n", subject)
	message += fmt.Sprintf("\r\n%s\r\n", body)

	err := smtp.SendMail(
		host+":"+port,
		nil,
		from,
		to,
		[]byte(message),
	)
	if err != nil {
		slog.Error("nmp", "err", err)
		return
	}
}
