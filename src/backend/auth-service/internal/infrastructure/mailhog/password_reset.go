package mailhog

func (s *Server) SendPasswordResetMail(to []string, resetLink string) error {
	subject := "[POUCH] change password request"

	body, err := s.renderTemplate("password_reset.template.html", resetLink)
	if err != nil {
		return err
	}

	return s.sendMail(subject, body, to)
}
