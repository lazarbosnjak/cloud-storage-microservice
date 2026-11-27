package mailhog

import (
	"math"
	"time"
)

type OtpTemplateData struct {
	Code             string
	ExpiresInMinutes int
}

func (s *Server) SendOtpMail(to []string, code string, expiresIn time.Duration) error {
	subject := "[POUCH] Your sign-in code"

	data := OtpTemplateData{
		Code:             code,
		ExpiresInMinutes: int(math.Ceil(expiresIn.Minutes())),
	}

	body, err := s.renderTemplate("otp.template.html", data)
	if err != nil {
		return err
	}

	return s.sendMail(subject, body, to)
}
