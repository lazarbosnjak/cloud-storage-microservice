package mailhog

func (s *Server) SendVerifyRegistrationMail(to []string, verificationLink string) error {
	subject := "[POUCH] Verify registration"

	templateName := "registration.template.html"
	body, err := s.renderTemplate(templateName, verificationLink)
	if err != nil {
		return err
	}

	err = s.sendMail(subject, body, to)
	if err != nil {
		return err
	}

	return nil
}
