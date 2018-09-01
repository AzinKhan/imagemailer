package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"time"

	"github.com/gorilla/mux"

	"github.com/jordan-wright/email"
)

const addr string = "booboophotomailer@gmail.com"
const passwd string = "booboo123!"

var toAddress string = "azink91@googlemail.com"

var serverPort string

func init() {
	flag.StringVar(&serverPort, "p", "8000", "Port for HTTP server")
	flag.StringVar(&toAddress, "t", toAddress, "Address to send email")
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
		log.Println("Sending email...")
		err := sendEmail(toAddress, bytes.NewReader(data))
		if err != nil {
			log.Printf("Could not send email: %+v", err)
		}
		log.Println("...done")
	}
}

func HandlePost(imChan imageChannel) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received post")
		err := r.ParseMultipartForm(10000000000)
		if err != nil {
			log.Println("Could not read data from file")
			w.WriteHeader(200)
		}
		datas := r.MultipartForm
		for _, headers := range datas.File {
			aux, err := headers[0].Open()
			if err != nil {
				log.Println(err)
				break
			}
			//name := headers[0].Filename
			file, err := ioutil.ReadAll(aux)
			if err != nil {
				log.Println(err)
				break
			}
			imChan <- file
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
	log.Println("Starting HTTP server..")
	log.Fatal(server.ListenAndServe())
}
