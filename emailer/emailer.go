package emailer

import (
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/smtp"
	"os"
	"os/exec"
	"path/filepath"
	"time"

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
	//var data [][]byte
	size := 0
	// Either append until memory limit reached or
	// until timeout
	for {
		counter := 0
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
				/*
						_, err := e.mail.Attach(bytes.NewReader(a.data), a.filename, a.content)
						if err != nil {
							log.Printf("Could not attach: %+v", err)
							continue
						}
					data = append(data, a.data)
					log.Printf("Length of data:\t%+v", len(data))
				*/
				// Write to file
				filename := fmt.Sprintf("image%d.jpg", counter)
				err := ioutil.WriteFile(filename, a.data, 0644)
				if err != nil {
					log.Println(err)
				} else {
					counter++
				}
				size += len(a.data)
				log.Printf("Total size: %+v", size)
				if size >= 20000000 {
					log.Println("Maximum attachment size reached")
					break AttachLoop
				}
			}
		}
		/*
			GIF, err := giffer.Giffer(data)
			if err != nil {
				log.Println(err)
			}
		*/
		cmd := exec.Command("ffmpeg", "-i", "image%d.jpg", "video.avi")
		err := cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
		cmd2 := exec.Command("ffmpeg", "-i", "video.avi", "out.gif")
		err = cmd2.Run()
		if err != nil {
			log.Fatal(err)
		}
		GIF, err := os.Open("out.gif")
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
		os.Remove("out.gif")
		os.Remove("video.avi")
		files, err := filepath.Glob("./*jpg")
		if err != nil {
			log.Println(err)
		}
		for _, file := range files {
			os.Remove(file)
		}
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
