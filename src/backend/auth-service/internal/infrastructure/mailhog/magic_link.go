package mailhog

func (s *Server) SendMagicLinkMail(to []string, magicLink string) error {
	subject := "[POUCH] sign in request"

	body, err := s.renderTemplate("magic_link.template.html", magicLink)
	if err != nil {
		return err
	}

	return s.sendMail(subject, body, to)
}
