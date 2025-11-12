package email

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
)

// Config holds email configuration
type Config struct {
	SMTPHost     string
	SMTPPort     string
	FromEmail    string
	FromName     string
	SMTPUsername string // Optional for auth
	SMTPPassword string // Optional for auth
}

// Service handles sending emails
type Service struct {
	config *Config
}

// NewService creates a new email service
func NewService(config *Config) *Service {
	return &Service{
		config: config,
	}
}

// Email represents an email message
type Email struct {
	To      string
	Subject string
	Body    string
	IsHTML  bool
}

// Send sends an email
func (s *Service) Send(email *Email) error {
	from := fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromEmail)

	var body bytes.Buffer
	body.WriteString(fmt.Sprintf("From: %s\r\n", from))
	body.WriteString(fmt.Sprintf("To: %s\r\n", email.To))
	body.WriteString(fmt.Sprintf("Subject: %s\r\n", email.Subject))

	if email.IsHTML {
		body.WriteString("MIME-version: 1.0;\r\n")
		body.WriteString("Content-Type: text/html; charset=\"UTF-8\";\r\n\r\n")
	} else {
		body.WriteString("Content-Type: text/plain; charset=\"UTF-8\";\r\n\r\n")
	}

	body.WriteString(email.Body)

	addr := s.config.SMTPHost + ":" + s.config.SMTPPort

	// For development (Mailpit), no auth is needed
	if s.config.SMTPUsername == "" {
		return smtp.SendMail(
			addr,
			nil,
			s.config.FromEmail,
			[]string{email.To},
			body.Bytes(),
		)
	}

	// For production with authentication
	auth := smtp.PlainAuth("", s.config.SMTPUsername, s.config.SMTPPassword, s.config.SMTPHost)
	return smtp.SendMail(
		addr,
		auth,
		s.config.FromEmail,
		[]string{email.To},
		body.Bytes(),
	)
}

// Template represents an email template
type Template struct {
	Subject  string
	BodyHTML string
	BodyText string
}

// SendFromTemplate sends an email using a template
func (s *Service) SendFromTemplate(to string, tmpl *Template, data interface{}) error {
	// Parse and execute HTML template
	htmlTmpl, err := template.New("email").Parse(tmpl.BodyHTML)
	if err != nil {
		return fmt.Errorf("failed to parse HTML template: %w", err)
	}

	var htmlBody bytes.Buffer
	if err := htmlTmpl.Execute(&htmlBody, data); err != nil {
		return fmt.Errorf("failed to execute HTML template: %w", err)
	}

	email := &Email{
		To:      to,
		Subject: tmpl.Subject,
		Body:    htmlBody.String(),
		IsHTML:  true,
	}

	return s.Send(email)
}
