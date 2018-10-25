package emailer

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/smtp"
	"time"

	"github.com/AzinKhan/giffer"
	"github.com/jordan-wright/email"
)

var addr string
var passwd string

type addressSlice []string

func (a *addressSlice) Set(value string) error {
	*a = append(*a, value)
	return nil
}

func (a *addressSlice) String() string {
	return fmt.Sprintf("%s", *a)
}

var toAddresses addressSlice
var serverPort string

func init() {
	flag.StringVar(&serverPort, "p", "8000", "Port for HTTP server")
	flag.StringVar(&addr, "addr", "", "Email address from which to send images")
	flag.StringVar(&passwd, "pass", "", "Password for account")
	flag.Var(&toAddresses, "t", "Addresses to send email")
}

type ImageChannel chan attachment

type Emailer struct {
	Mail     *email.Email
	ImChan   ImageChannel
	passwd   string
	sendFunc func(Sender, *emailConfig) error
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

type Sender interface {
	Send(string, smtp.Auth) error
}

type emailConfig struct {
	address string
	auth    smtp.Auth
}

type Attachment struct {
	Data        *bytes.Buffer
	Filename    string
	ContentType string
}

type OutputChan chan *Attachment

func NewEmailer(c Creds) Emailer {
	var e Emailer
	e.Mail = email.NewEmail()
	e.Mail.From = c.From
	e.Mail.To = c.To
	e.passwd = c.Password
	e.sendFunc = send
	return e
}

func (e *Emailer) Send() error {
	log.Printf("Sending email to %+v", e.Mail.To)
	e.Mail.Subject = fmt.Sprintf("Motion detected! Time: %+v", time.Now())
	now := fmt.Sprintf("Email sent: %+v", time.Now())
	e.Mail.Text = []byte(now)
	emailAuth := &emailConfig{
		address: "smtp.gmail.com:587",
		auth: smtp.PlainAuth(
			"", e.Mail.From, e.passwd, "smtp.gmail.com",
		),
	}
	return e.sendFunc(e.Mail, emailAuth)
}

func send(s Sender, a *emailConfig) error {
	return s.Send(a.address, a.auth)
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

func MakeGIFAttachment(data [][]byte) (*Attachment, error) {
	GIF, err := giffer.Giffer(data)
	if err != nil {
		return &Attachment{}, err
	}
	a := Attachment{
		Data:        GIF,
		Filename:    fmt.Sprintf("%+v.gif", time.Now()),
		ContentType: "image/gif",
	}
	return &a, nil
}

func BufferImages(ImChan ImageChannel, outChan OutputChan) {
	t := time.NewTimer(20 * time.Second)
	var data [][]byte
	size := 0
	// Either append until memory limit reached or
	// until timeout
	for {
		select {
		case <-t.C:
			if size > 0 {
				log.Println("Timeout reached")
				GIF, err := MakeGIFAttachment(data)
				if err != nil {
					log.Println(err)
					continue
				}
				outChan <- GIF
				size = 0
				data = nil
				t.Reset(20 * time.Second)
			} else {
				log.Println("No images received, resetting timer.")
				t.Reset(20 * time.Second)
			}
		case a := <-ImChan:
			t.Reset(20 * time.Second)
			log.Printf("Collecting attachment %+v", a.Filename)
			data = append(data, a.Data)
			log.Printf("Length of data:\t%+v", len(data))
			// Write to file
			size += len(a.Data)
			log.Printf("Total size: %+v", size)
			if size >= 2000000 {
				log.Println("Maximum attachment size reached")
				GIF, err := MakeGIFAttachment(data)
				if err != nil {
					log.Println(err)
					continue
				}
				outChan <- GIF
				size = 0
				data = nil
				t.Reset(20 * time.Second)
			}
		}
	}
}

func Email(outChan OutputChan, creds Creds) {
	for {
		attachment := <-outChan
		log.Printf("Received file to email %+v", attachment.Filename)
		e := NewEmailer(creds)
		_, err := e.Mail.Attach(attachment.Data, attachment.Filename, attachment.ContentType)
		if err != nil {
			log.Println(err)
			continue
		}
		err = e.Send()
		if err != nil {
			log.Printf("Could not send email: %+v", err)
		} else {
			log.Println("...done")
		}
	}
}
