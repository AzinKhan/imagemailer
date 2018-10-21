package emailer

import (
	"fmt"
	"log"
	"net/smtp"
	"testing"
)

const (
	from     = "test@gmail.com"
	password = "password"
)

var sendAddresses = []string{"booboo@gmail.com", "booboo2@gmail.com"}

func TestNewEmailer(t *testing.T) {
	creds := Creds{
		To:       sendAddresses,
		From:     from,
		Password: password,
	}
	emailer := NewEmailer(creds)
	if emailer.passwd != password {
		t.Fail()
	}
	if emailer.Mail.From != from {
		t.Fail()
	}
	for i, address := range sendAddresses {
		if address != emailer.Mail.To[i] {
			t.Fail()
		}
	}
}

func TestSend(t *testing.T) {
	authChan := make(chan *emailConfig, 1)
	defer close(authChan)
	send := func(s Sender, a *emailConfig) error {
		authChan <- a
		return nil
	}
	creds := Creds{
		To:       sendAddresses,
		From:     from,
		Password: password,
	}
	emailer := NewEmailer(creds)
	emailer.sendFunc = send
	err := emailer.Send()
	if err != nil {
		log.Printf("%+v", err)
		t.Fail()
	}
	expectedAuth := smtp.PlainAuth(
		"", from, password, "smtp.gmail.com",
	)
	auth := <-authChan
	expectedAddr := "smtp.gmail.com:587"
	if auth.address != expectedAddr {
		log.Printf("Expected: %+v", expectedAddr)
		log.Printf("Received: %+v", auth.address)
		t.Fail()
	}
	// Cast to string due to unexported fields
	result := fmt.Sprintf("%+v", auth.auth)
	expected := fmt.Sprintf("%+v", expectedAuth)
	if result != expected {
		log.Printf("Expected: %+v", expectedAuth)
		log.Printf("Received: %+v", auth.auth)
		t.Fail()
	}
}
