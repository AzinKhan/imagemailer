package emailer

import (
	"fmt"
	"log"
	"net/smtp"

	"github.com/jordan-wright/email"
)

type Mailer struct {
	address    string
	auth       smtp.Auth
	username   string
	recipients []string
}

func NewMailer(username, password, host, address string, recipients ...string) *Mailer {
	return &Mailer{
		address:    address,
		auth:       smtp.PlainAuth("", username, password, host),
		username:   username,
		recipients: recipients,
	}
}

func (m *Mailer) Send(msg Email) error {
	if len(m.recipients) == 0 {
		return nil
	}

	e := email.NewEmail()
	e.From = m.username
	e.To = m.recipients
	e.Text = msg.Body
	e.Subject = msg.Subject
	_, err := e.Attach(msg.Attachment.Data, msg.Attachment.Filename, msg.Attachment.ContentType)
	if err != nil {
		return fmt.Errorf("attaching: %w", err)
	}
	log.Printf("Sending email with subject %s to %v\n", msg.Subject, m.recipients)
	return e.Send(m.address, m.auth)
}
