package mailutils_test

import (
	"errors"
	"html/template"
	"strings"
	"testing"

	"github.com/gurch101/gowebutils/pkg/mailutils"
)

func TestEmailer_Send_Success(t *testing.T) {
	t.Parallel()
	// Setup test data
	templates := map[string]*template.Template{
		"testTemplate": template.Must(template.New("testTemplate").Parse(`
			{{ define "subject" }}Test Subject{{ end }}
			{{ define "plainBody" }}Plain Body {{.Test}}{{ end }}
			{{ define "htmlBody" }}<p>HTML Body</p>{{ end }}
		`)),
	}

	emailer := mailutils.NewMockMailer(mailutils.WithEmailTemplateMap(templates))

	// Test data
	recipient := "recipient@example.com"
	templateName := "testTemplate"
	data := map[string]string{"Test": "TestValue"}

	// Execute the method
	emailer.Send(recipient, templateName, data)
	msg := emailer.MessageToString(0)

	if !strings.Contains(msg, "Subject: Test Subject") {
		t.Errorf("Expected subject to be 'Test Subject', got %s", msg)
	}

	if !strings.Contains(msg, "Plain Body TestValue") {
		t.Errorf("Expected plain body to be 'Plain Body', got %s", msg)
	}

	if !strings.Contains(msg, "<p>HTML Body</p>") {
		t.Errorf("Expected HTML body to be '<p>HTML Body</p>', got %s", msg)
	}
}

func TestEmailer_Send_TemplateNotFound(t *testing.T) {
	t.Parallel()
	// Setup test data
	templates := map[string]*template.Template{}
	emailer := mailutils.NewMockMailer(mailutils.WithEmailTemplateMap(templates))

	// Test data
	recipient := "recipient@example.com"
	templateName := "nonExistentTemplate"
	data := map[string]string{}

	// Execute the method
	emailer.Send(recipient, templateName, data)

	// Assertions
	if !errors.Is(emailer.Error, mailutils.ErrTemplateNotFound) {
		t.Errorf("Expected ErrTemplateNotFound, got %v", emailer.Error)
	}
}
