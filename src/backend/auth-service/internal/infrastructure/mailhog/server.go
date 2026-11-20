package mailhog

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/smtp"

	"auth-service/internal/startup/config"
)

//go:embed templates/*.template.html
var templateFS embed.FS

var templates = template.Must(template.ParseFS(templateFS, "templates/*.template.html"))

var ErrExecutingTemplate = errors.New("error executing html template")

type Server struct {
	config config.Config
}

func NewServer(config config.Config) *Server {
	return &Server{
		config: config,
	}
}

func (s *Server) sendMail(subject, body string, to []string) error {
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
	message += "Content-Type: text/html; charset=\"UTF-8\"\r\n"
	message += fmt.Sprintf("\r\n%s\r\n", body)

	err := smtp.SendMail(
		host+":"+port,
		nil,
		from,
		to,
		[]byte(message),
	)
	if err != nil {
		slog.Error("error sending mail", "err", err)
		return err
	}

	return nil
}

func (s *Server) renderTemplate(name string, data any) (string, error) {
	var body bytes.Buffer
	if err := templates.ExecuteTemplate(&body, name, data); err != nil {
		return "", fmt.Errorf("%w %q: %w", ErrExecutingTemplate, name, err)
	}

	return body.String(), nil
}
