package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/AzinKhan/giffer"
	"github.com/AzinKhan/imagemailer/emailer"

	"github.com/gorilla/mux"
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

type Attachment struct {
	Data        *bytes.Buffer
	Filename    string
	ContentType string
}

type OutputChan chan *Attachment

func init() {
	flag.StringVar(&serverPort, "p", "8000", "Port for HTTP server")
	flag.Var(&toAddresses, "t", "Addresses to send email")
	flag.StringVar(&addr, "addr", "", "Email address from which to send images")
	flag.StringVar(&passwd, "pass", "", "Password for account")
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

func BufferImages(ImChan emailer.ImageChannel, outChan OutputChan) {
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

func Email(outChan OutputChan, creds emailer.Creds) {
	for {
		attachment := <-outChan
		log.Printf("Received file to email %+v", attachment.Filename)
		e := emailer.NewEmailer(creds)
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

func main() {
	flag.Parse()
	emailChan := make(emailer.ImageChannel)
	outByteChan := make(OutputChan)
	address := fmt.Sprintf("0.0.0.0:%v", serverPort)
	router := mux.NewRouter()
	router.HandleFunc("/", emailer.HandlePost(emailChan))
	server := &http.Server{
		Addr:    address,
		Handler: router,
	}
	log.Printf("Email clients are: %+v", toAddresses)
	credentials := emailer.Creds{
		To:       toAddresses,
		From:     addr,
		Password: passwd,
	}
	go Email(outByteChan, credentials)
	go BufferImages(emailChan, outByteChan)
	log.Println("Starting HTTP server..")
	log.Fatal(server.ListenAndServe())
}
