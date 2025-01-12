package testutils

type MockMailer struct {
	SentEmails []map[string]any
}

func NewMockMailer() *MockMailer {
	return &MockMailer{
		SentEmails: []map[string]any{},
	}
}

func (m *MockMailer) Send(recipient, templateName string, data map[string]string) {
	email := map[string]any{
		"recipient":    recipient,
		"templateName": templateName,
		"data":         data,
	}
	m.SentEmails = append(m.SentEmails, email)
}
