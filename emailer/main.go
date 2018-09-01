package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"time"

	"github.com/gorilla/mux"

	"github.com/jordan-wright/email"
)

var addr string
var passwd string

var toAddress string = "azink91@googlemail.com"

var serverPort string

func init() {
	flag.StringVar(&serverPort, "p", "8000", "Port for HTTP server")
	flag.StringVar(&toAddress, "t", toAddress, "Address to send email")
	flag.StringVar(&addr, "addr", "", "Email address from which to send images")
	flag.StringVar(&passwd, "pass", "", "Password for account")
}

func sendEmail(toAddr string, attachment io.Reader) error {
	e := email.NewEmail()
	e.To = []string{toAddr}
	e.From = addr
	e.Subject = "Motion detected!"
	now := fmt.Sprintf("Photo received: %+v", time.Now())
	e.Text = []byte(now)
	e.Attach(attachment, fmt.Sprintf("image%+v.jpg", time.Now().Unix()), "image/jpeg")
	return e.Send("smtp.gmail.com:587", smtp.PlainAuth("", addr, passwd, "smtp.gmail.com"))
}

type imageChannel chan []byte

func waitForEmail(eChan imageChannel) {
	for {
		data := <-eChan
		err := sendEmail(toAddress, bytes.NewReader(data))
		if err != nil {
			log.Printf("Could not send email: %+v", err)
		}
	}
}

func HandlePost(imChan imageChannel) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(100000000)
		file, _, err := r.FormFile("image")
		if err != nil {
			w.WriteHeader(403)
		} else {
			defer file.Close()
			var data []byte
			file.Read(data)
			imChan <- data
			w.WriteHeader(200)
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
	go waitForEmail(emailChan)
	log.Fatal(server.ListenAndServe())
}
