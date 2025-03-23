# Mail

[`godoc`](https://pkg.go.dev/github.com/gurch101/gowebutils/pkg/mailutils)

The `mailutils` package provides a simple way to send templated emails asynchronously via goroutines.

### Initialization

To initialize the mailer, follow these steps:

1. When creating the `App` instance, embed your email templates and pass them to the `app.NewApp` function.

```go
//go:embed templates/email
var emailTemplates embed.FS

func main() {
  app, err := app.NewApp(
    app.WithEmailTemplates(emailTemplates),
  )
  if err != nil {
    log.Fatal("Failed to initialize app:", err)
  }
}
```

2. Configure SMTP Settings

Start your app with the following environment variables:

```bash
export SMTP_HOST="my.smtp.host.com"
export SMTP_PORT="587"
export SMTP_USERNAME="myusername"
export SMTP_PASSWORD="mypassword"
export SMTP_FROM="admin@myapp.com" # Default "From" email address
```

### Usage

Once initialized, the mailer is accessible via the `App` instance. Use the `Send` method to send emails asynchronously.

```go
app.Mailer.Send(
  "recipient@example.com", // Recipient email address
  "mytemplatename.go.tmpl", // Email template name relative to the embedded filesystem directory
  map[string]string{ // Template data
    "name": "John Doe",
  },
)
```

### Email Templates

The mailer uses Go's `html/template` package to render email content. Each template must define three sections:

1. `subject`: The email subject.
2. `plainBody`: The plain text version of the email body.
3. `htmlBody`: The HTML version of the email body.

#### Example Template

```go
{{define "subject"}}Hello {{.name}}!{{end}}
{{define "plainBody"}}Hello {{.name}}, this is a plain text email.{{end}}
{{define "htmlBody"}}<h1>Hello {{.name}}</h1>, <p>this is an html email.</p>{{end}}
```

### Testing

TODO
