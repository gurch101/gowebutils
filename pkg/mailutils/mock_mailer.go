package mailutils

import (
	"bytes"
	"embed"
	"html/template"

	"github.com/gurch101/gowebutils/pkg/templateutils"
	"gopkg.in/gomail.v2"
)

type MockDialer struct {
	Messages []*gomail.Message
}

func (d *MockDialer) DialAndSend(messages ...*gomail.Message) error {
	d.Messages = append(d.Messages, messages...)

	return nil
}

type MockMailer struct {
	SentEmails []map[string]any
	mailer     *Emailer
	Error      error
}

type Option func(options *options) error

type options struct {
	emailTemplateMap map[string]*template.Template
}

func WithEmailTemplates(emailTemplates *embed.FS) Option {
	return func(options *options) error {
		if emailTemplates == nil {
			return nil
		}

		options.emailTemplateMap = templateutils.LoadTemplates(*emailTemplates)

		return nil
	}
}

func WithEmailTemplateMap(emailTemplateMap map[string]*template.Template) Option {
	return func(options *options) error {
		options.emailTemplateMap = emailTemplateMap

		return nil
	}
}

func NewMockMailer(opts ...Option) *MockMailer {
	var options options
	for _, opt := range opts {
		err := opt(&options)
		if err != nil {
			return nil
		}
	}

	mailer := InitMailerWithDialer(
		&MockDialer{Messages: []*gomail.Message{}},
		"admin@example.com",
		options.emailTemplateMap,
	)

	return &MockMailer{
		SentEmails: []map[string]any{},
		mailer:     mailer,
		Error:      nil,
	}
}

func (m *MockMailer) Dialer() *MockDialer {
	val, ok := m.mailer.dialer.(*MockDialer)
	if !ok {
		return nil
	}

	return val
}

func (m *MockMailer) MessageToString(index int) string {
	var buf bytes.Buffer

	dialer := m.Dialer()
	// Write the message to the buffer
	_, err := dialer.Messages[index].WriteTo(&buf)
	if err != nil {
		return ""
	}

	return buf.String()
}

func (m *MockMailer) Send(recipient, templateName string, data map[string]string) {
	email := map[string]any{
		"recipient":    recipient,
		"templateName": templateName,
		"data":         data,
	}

	m.SentEmails = append(m.SentEmails, email)
	if m.mailer != nil {
		m.Error = m.mailer.sendInternal(recipient, templateName, data)
	}
}
