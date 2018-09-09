package emailer

import (
	"fmt"
	"imagemailer/giffer"
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
	mail   *email.Email
	ImChan ImageChannel
	passwd string
}

type attachment struct {
	data     []byte
	filename string
	content  string
}

type Creds struct {
	To       []string
	From     string
	Password string
}

func NewEmailer(c Creds) Emailer {
	var e Emailer
	e.mail = email.NewEmail()
	e.mail.From = c.From
	e.mail.To = c.To
	e.passwd = c.Password
	return e
}

func (e *Emailer) Send() error {
	log.Printf("Sending email to %+v", e.mail.To)
	e.mail.Subject = fmt.Sprintf("Motion detected! Time: %+v", time.Now())
	now := fmt.Sprintf("Email sent: %+v", time.Now())
	e.mail.Text = []byte(now)
	return e.mail.Send("smtp.gmail.com:587", smtp.PlainAuth("", e.mail.From, e.passwd, "smtp.gmail.com"))
}

func (e *Emailer) Run() {
	t := time.NewTimer(20 * time.Second)
	var data [][]byte
	size := 0
	// Either append until memory limit reached or
	// until timeout
	for {
	AttachLoop:
		for {
			select {
			case <-t.C:
				if size > 0 {
					log.Println("Timeout reached")
					break AttachLoop
				} else {
					log.Println("No images received, resetting timer.")
					t.Reset(20 * time.Second)
				}
			case a := <-e.ImChan:
				t.Reset(20 * time.Second)
				log.Printf("Collecting attachment %+v", a.filename)
				data = append(data, a.data)
				log.Printf("Length of data:\t%+v", len(data))
				// Write to file
				size += len(a.data)
				log.Printf("Total size: %+v", size)
				if size >= 2000000 {
					log.Println("Maximum attachment size reached")
					break AttachLoop
				}
			}
		}
		GIF, err := giffer.Giffer(data)
		if err != nil {
			log.Println(err)
		}
		_, err = e.mail.Attach(GIF, "test.gif", "image/gif")
		if err != nil {
			log.Println(err)
		}
		log.Printf("Files have total size %v", size)
		err = e.Send()
		if err != nil {
			log.Printf("Could not send email: %+v", err)
		} else {
			log.Println("...done")
		}
		// Clear attachments
		log.Printf("Clearing %+v attachments", len(e.mail.Attachments))
		e.mail.Attachments = nil
		log.Printf("%+v attachments remaining", len(e.mail.Attachments))
		size = 0
		data = nil
		t.Reset(20 * time.Second)
	}
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
					data:     file,
					filename: name,
					content:  "image/jpeg",
				}
				imChan <- newFile
			}
		}
	}
}
