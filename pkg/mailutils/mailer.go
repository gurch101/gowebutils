// The mailer package contains a Mailer struct and associated methods for sending
// emails
package mailutils

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"text/template"
	"time"

	"github.com/gurch101/gowebutils/pkg/parser"
	"github.com/gurch101/gowebutils/pkg/threads"
	"gopkg.in/gomail.v2"
)

// ErrTemplateNotFound is returned when the template is not found.
var ErrTemplateNotFound = errors.New("template not found")

// ErrTemplateExecution is returned when there is an error executing the template.
var ErrTemplateExecution = errors.New("error executing mail template")

const retryInterval = 500 * time.Millisecond

// MailSender is an interface for sending emails.
type MailSender interface {
	Send(recipient, templateName string, data map[string]string) error
}

// Mailer is a struct for sending emails.
type Mailer struct {
	dialer    *gomail.Dialer
	sender    string
	templates map[string]*template.Template
}

// New initializes a new Mailer instance.
func NewMailer(
	host string,
	port int,
	username, password, sender string,
	templates map[string]*template.Template,
) *Mailer {
	dialer := gomail.NewDialer(host, port, username, password)

	return &Mailer{dialer: dialer, sender: sender, templates: templates}
}

// InitMailer initializes a new Mailer instance from SMTP_HOST, SMTP_PORT,
// SMTP_USERNAME, SMTP_PASSWORD, and SMTP_FROM environment variables.
func InitMailer(templates map[string]*template.Template) *Mailer {
	return NewMailer(
		parser.ParseEnvStringPanic("SMTP_HOST"),
		parser.ParseEnvIntPanic("SMTP_PORT"),
		parser.ParseEnvStringPanic("SMTP_USERNAME"),
		parser.ParseEnvStringPanic("SMTP_PASSWORD"),
		parser.ParseEnvStringPanic("SMTP_FROM"),
		templates,
	)
}

// Send sends an email from a template using the provided data.
func (m *Mailer) Send(recipient, templateName string, data map[string]string) {
	threads.Background(func() {
		err := m.sendInternal(recipient, templateName, data)
		if err == nil {
			slog.Error("failed to send invite email", "error", err)
		}
	})
}

func (m *Mailer) sendInternal(recipient, templateName string, data map[string]string) error {
	var err error

	tmpl, ok := m.templates[templateName]
	if !ok {
		return ErrTemplateNotFound
	}

	subject := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(subject, "subject", data); err != nil {
		return fmt.Errorf("%w: %w", ErrTemplateExecution, err)
	}

	plainBody := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(plainBody, "plainBody", data); err != nil {
		return fmt.Errorf("%w: %w", ErrTemplateExecution, err)
	}

	htmlBody := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(htmlBody, "htmlBody", data); err != nil {
		return fmt.Errorf("%w: %w", ErrTemplateExecution, err)
	}

	msg := gomail.NewMessage()
	msg.SetHeader("From", m.sender)
	msg.SetHeader("To", recipient)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", plainBody.String())
	msg.AddAlternative("text/html", htmlBody.String())

	// Try sending the email up to three times before aborting and returning the final
	// error. We sleep for 500 milliseconds between each attempt.
	for i := 1; i <= 3; i++ {
		err := m.dialer.DialAndSend(msg)
		if nil == err {
			return nil
		}
		// If it didn't work, sleep for a short time and retry.
		time.Sleep(retryInterval)
	}

	return err
}
