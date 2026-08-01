package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"time"

	mail "github.com/xhit/go-simple-mail/v2"
)

//go:embed templates
var emailTemplatesFS embed.FS

type emailMessage struct {
	from    string
	to      string
	subject string
	tmpl    string
	data    any
}

func (app *application) getTemplate(name, tmpl string, isPlain bool) (*template.Template, error) {

	cacheKey := fmt.Sprintf("%s-%t", tmpl, isPlain)

	// template caching for prod
	app.emailTemplateMu.RLock()
	t, ok := app.emailTemplateCache[cacheKey]
	app.emailTemplateMu.RUnlock()
	if app.config.env == "production" && ok {

		return t, nil
	}

	var templateToRender string
	if isPlain {

		templateToRender = fmt.Sprintf("templates/%s.plain.tmpl", tmpl)

	} else {

		templateToRender = fmt.Sprintf("templates/%s.html.tmpl", tmpl)
	}

	t, err := template.New(name).ParseFS(emailTemplatesFS, templateToRender)
	if err != nil {

		return nil, err
	}

	app.emailTemplateMu.Lock()
	app.emailTemplateCache[cacheKey] = t
	app.emailTemplateMu.Unlock()

	return t, nil
}

func (app *application) listenForMail() {

	// KeepAlive = false, fresh connection for each email, no stale connection error but slower; with true it retries
	server := mail.NewSMTPClient()
	server.Host = app.config.smtp.host
	server.Port = app.config.smtp.port
	server.Username = app.config.smtp.username
	server.Password = app.config.smtp.password
	server.Encryption = mail.EncryptionSTARTTLS
	server.KeepAlive = true
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	var smtpClient *mail.SMTPClient

	// try to send the email 3 times
	app.mailer.Listen(func(msg emailMessage) {

		var err error
		for i := 1; i <= 3; i++ {

			if smtpClient == nil {

				smtpClient, err = server.Connect()
				if err != nil {

					app.errorLog.Println("SMTP connection failed:", err)
					time.Sleep(2 * time.Second)
					continue
				}
			} else if err = smtpClient.Noop(); err != nil {

				// Connection went stale, reconnect
				smtpClient, err = server.Connect()
				if err != nil {

					app.errorLog.Println("SMTP reconnection failed:", err)
					smtpClient = nil
					time.Sleep(2 * time.Second)
					continue
				}
			}

			err = app.sendEmailMessage(msg, smtpClient)
			if err != nil {

				app.errorLog.Printf("Failed to send email (attempt %d): %v", i, err)
				smtpClient = nil // Force reconnect on next iteration
				time.Sleep(2 * time.Second)
				continue
			}

			break // Sent successfully
		}

		if err != nil {

			app.errorLog.Printf("Email to %s dropped after 3 failed attempts: %v", msg.to, err)
		}
	}, app.wg)
}

func (app *application) SendEmail(from, to, subject, tmpl string, data any) error {

	return app.mailer.Send(emailMessage{
		from:    from,
		to:      to,
		subject: subject,
		tmpl:    tmpl,
		data:    data,
	})
}

func (app *application) sendEmailMessage(msg emailMessage, smtpClient *mail.SMTPClient) error {

	t, err := app.getTemplate("email-html", msg.tmpl, false)
	if err != nil {

		return err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", msg.data); err != nil {

		return err
	}
	formattedMessage := tpl.String()

	t, err = app.getTemplate("email-plain", msg.tmpl, true)
	if err != nil {

		return err
	}

	tpl.Reset()
	if err = t.ExecuteTemplate(&tpl, "body", msg.data); err != nil {

		return err
	}
	plainMessage := tpl.String()

	email := mail.NewMSG()
	email.SetFrom(msg.from).AddTo(msg.to).SetSubject(msg.subject)
	email.SetBody(mail.TextHTML, formattedMessage)
	email.AddAlternative(mail.TextPlain, plainMessage)

	err = email.Send(smtpClient)
	if err != nil {

		return err
	}

	return nil
}
