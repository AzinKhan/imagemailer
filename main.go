package main

import (
	"flag"
	"fmt"
	"imagemailer/emailer"
	"log"
	"net/http"

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

func init() {
	flag.StringVar(&serverPort, "p", "8000", "Port for HTTP server")
	flag.Var(&toAddresses, "t", "Addresses to send email")
	flag.StringVar(&addr, "addr", "", "Email address from which to send images")
	flag.StringVar(&passwd, "pass", "", "Password for account")
}

func main() {
	flag.Parse()
	emailChan := make(emailer.ImageChannel)
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
	emailer := emailer.NewEmailer(credentials)
	emailer.ImChan = emailChan
	go emailer.Run()
	log.Println("Starting HTTP server..")
	log.Fatal(server.ListenAndServe())
}
