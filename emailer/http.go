package emailer

import (
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
)

const maxFormSize = 10000000

func assembleFile(h []*multipart.FileHeader) ([]byte, error) {
	aux, err := h[0].Open()
	if err != nil {
		return nil, err
	}
	file, err := ioutil.ReadAll(aux)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func getForm(r *http.Request) (*multipart.Form, error) {
	err := r.ParseMultipartForm(maxFormSize)
	if err != nil {
		return nil, fmt.Errorf("getting multipart form: %w", err)
	}
	return r.MultipartForm, nil
}

func NewUploadHandler(p Processor) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		form, err := getForm(r)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		for _, headers := range form.File {
			file, err := assembleFile(headers)
			if err != nil {
				log.Printf("Could not construct file from multipart request: %v\n", err)
				continue
			}
			p.Process(file)
		}
	}
}
