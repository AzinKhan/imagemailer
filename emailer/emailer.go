package emailer

import (
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/smtp"
	"time"

	"github.com/jordan-wright/email"
)

type ImageChannel chan attachment

type Emailer struct {
	Mail   *email.Email
	ImChan ImageChannel
	passwd string
}

type attachment struct {
	Data     []byte
	Filename string
	Content  string
}

type Creds struct {
	To       []string
	From     string
	Password string
}

func NewEmailer(c Creds) Emailer {
	var e Emailer
	e.Mail = email.NewEmail()
	e.Mail.From = c.From
	e.Mail.To = c.To
	e.passwd = c.Password
	return e
}

func (e *Emailer) Send() error {
	log.Printf("Sending email to %+v", e.Mail.To)
	e.Mail.Subject = fmt.Sprintf("Motion detected! Time: %+v", time.Now())
	now := fmt.Sprintf("Email sent: %+v", time.Now())
	e.Mail.Text = []byte(now)
	return e.Mail.Send("smtp.gmail.com:587", smtp.PlainAuth("", e.Mail.From, e.passwd, "smtp.gmail.com"))
}

func AssembleFile(h []*multipart.FileHeader) ([]byte, string, error) {
	aux, err := h[0].Open()
	if err != nil {
		return nil, "", err
	}
	name := h[0].Filename
	file, err := ioutil.ReadAll(aux)
	return file, name, err
}

func GetForm(r *http.Request) (*multipart.Form, error) {
	err := r.ParseMultipartForm(10000000)
	if err != nil {
		return nil, err
	}
	return r.MultipartForm, nil
}

func HandlePost(imChan ImageChannel) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received post")
		datas, err := GetForm(r)
		if err != nil {
			log.Println("Could not get multipart form")
		}
		for _, headers := range datas.File {
			file, name, err := AssembleFile(headers)
			if err != nil {
				log.Println("Could not construct file from multipart request")
			} else {
				newFile := attachment{
					Data:     file,
					Filename: name,
					Content:  "image/jpeg",
				}
				imChan <- newFile
			}
		}
	}
}
