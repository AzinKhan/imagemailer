package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/smtp"
	"time"

	"github.com/gorilla/mux"
	"github.com/jordan-wright/email"
)

const addr string = "booboophotomailer@gmail.com"
const passwd string = "booboo123!"

var toAddress string
var serverPort string

func init() {
	flag.StringVar(&serverPort, "p", "8000", "Port for HTTP server")
	flag.StringVar(&toAddress, "t", toAddress, "Address to send email")
}

type imageChannel chan attachment

type Emailer struct {
	attachments []attachment
	mail        *email.Email
	imChan      imageChannel
	passwd      string
}

type attachment struct {
	data     []byte
	filename string
	content  string
}

type Creds struct {
	to       []string
	from     string
	password string
}

func NewEmailer(c Creds) Emailer {
	var e Emailer
	e.mail = email.NewEmail()
	e.mail.From = c.from
	e.mail.To = c.to
	e.passwd = c.password
	return e
}

func (e *Emailer) Attach() error {
	for _, a := range e.attachments {
		_, err := e.mail.Attach(bytes.NewReader(a.data), a.filename, a.content)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Emailer) Send() error {
	e.mail.Subject = "Motion detected!"
	now := fmt.Sprintf("Email sent: %+v", time.Now())
	e.mail.Text = []byte(now)
	return e.mail.Send("smtp.gmail.com:587", smtp.PlainAuth("", e.mail.From, e.passwd, "smtp.gmail.com"))
}

func (e *Emailer) Run() {
	t := time.NewTimer(20 * time.Second)
	for {
		// Either append until memory limit reached or
		// until timeout
		size := 0
	AttachLoop:
		for {
			select {
			case <-t.C:
				if size > 0 {
					log.Printf("Timeout reached, attaching files with total size %v", size)
					break AttachLoop
				} else {
					log.Println("No images received, resetting timer.")
					t.Reset(20 * time.Second)
				}
			case a := <-e.imChan:
				t.Reset(20 * time.Second)
				log.Println("Collecting attachment")
				e.attachments = append(e.attachments, a)
				size += len(a.data)
				log.Printf("Total size: %+v", size)
				if size >= 20000000 {
					log.Println("Maximum attachment size reached")
					break AttachLoop
				}
			}
		}

		err := e.Attach()
		if err != nil {
			log.Println("Could not attach file")
			continue
		}
		log.Println("Sending email")
		err = e.Send()
		if err != nil {
			log.Printf("Could not send email: %+v", err)
			continue
		} else {
			log.Println("...done")
		}
		// Clear attachments
		e.attachments = []attachment{}
		size = 0
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
	err := r.ParseMultipartForm(10000000000)
	if err != nil {
		return nil, err
	}
	return r.MultipartForm, nil
}

func HandlePost(imChan imageChannel) func(w http.ResponseWriter, r *http.Request) {
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
					filename: name + ".jpg",
					content:  "image/jpeg",
				}
				imChan <- newFile
			}
		}
	}
}

func main() {
	flag.Parse()
	emailChan := make(imageChannel)
	address := fmt.Sprintf("0.0.0.0:%v", serverPort)
	router := mux.NewRouter()
	router.HandleFunc("/", HandlePost(emailChan))
	server := &http.Server{
		Addr:    address,
		Handler: router,
	}
	credentials := Creds{
		to:       []string{"azink91@googlemail.com"},
		from:     addr,
		password: passwd,
	}
	emailer := NewEmailer(credentials)
	emailer.imChan = emailChan
	go emailer.Run()
	log.Println("Starting HTTP server..")
	log.Fatal(server.ListenAndServe())
}
