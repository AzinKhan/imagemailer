package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/AzinKhan/imagemailer/emailer"

	"github.com/gorilla/mux"
)

type addressSlice []string

func (a *addressSlice) Set(value string) error {
	*a = append(*a, value)
	return nil
}

func (a *addressSlice) String() string {
	return fmt.Sprintf("%s", *a)
}

var toAddresses addressSlice
var addr, passwd, serverPort, host, username string

func main() {
	flag.StringVar(&serverPort, "p", "8000", "Port for HTTP server")
	flag.Var(&toAddresses, "t", "Addresses to send email")
	flag.StringVar(&addr, "addr", "smtp.gmail.com:587", "Email address from which to send images")
	flag.StringVar(&username, "user", "", "Email username")
	flag.StringVar(&passwd, "pass", "", "Password for account")
	flag.StringVar(&host, "host", "smtp.gmail.com", "Email host")
	flag.Parse()

	ctx := context.Background()

	mailer := emailer.NewMailer(username, passwd, host, addr, toAddresses...)
	processor := emailer.NewImageProcessor(ctx, mailer)

	router := mux.NewRouter()
	router.HandleFunc("/", emailer.NewUploadHandler(processor))

	server := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%v", serverPort),
		Handler: router,
	}

	log.Printf("Email clients are: %+v", toAddresses)

	log.Printf("Starting HTTP server on %s\n", server.Addr)
	log.Fatal(server.ListenAndServe())
}
