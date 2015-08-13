package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
)

type Request struct {
	Filename string
	Mime     string
	Length   int
}

type Response struct {
	Url      string `json:"url"`
	Timeout  int    `json:"timeout"`
	Filename string `json:"filename"`
}

// These configuration variables will be read from command line flags later
var (
	address   = "localhost:3030"
	directory = "data"
)

func writeError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	fmt.Fprintf(w, "%s\r\n", message)
}

func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	} else if os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

func respond(w http.ResponseWriter, req Request) error {
	if req.Filename == "" && req.Mime == "" {
		writeError(w, 400, "Either a filename or MIME-type needs to be specified")
		return errors.New("Neither filename nor MIME-type supplied")
	}

	if req.Length <= 0 {
		writeError(w, 400, "Valid length needs to be specified.")
		return errors.New("No length specified")
	}

	if req.Filename == "" {
		writeError(w, 500, "Filename generation not implemented yet. Please supply a filename.")
		return errors.New("Filename generation not implemented")
	}

	if exists(fmt.Sprintf("%s/%s", directory, req.Filename)) {
		writeError(w, 409, "File with that name already exists.")
		return errors.New(fmt.Sprintf("File %s/%s already exists", directory, req.Filename))
	}

	// timeout not implemented yet
	resp := Response{Url: fmt.Sprintf("http://%s/%s", address, req.Filename), Filename: req.Filename, Timeout: 0}
	json.NewEncoder(w).Encode(resp)

	return nil
}

func transferRequestHandler(w http.ResponseWriter, req *http.Request) {
	var r Request

	// Need to handle OPTIONS properly
	if req.Method != "POST" {
		w.Header().Add("Allow", "POST")
		w.WriteHeader(405)
		return
	}

	// Need to check content-type
	err := json.NewDecoder(req.Body).Decode(&r)
	if err != nil {
		writeError(w, 400, "The JSON data you have sent is invalid.")
		return
	}

	err = respond(w, r)
	if err != nil {
		log.Println(err)
	}

	fmt.Printf("%v\n", r)
	w.WriteHeader(204)
}

func main() {
	fmt.Printf("%s => %s\n", address, directory)
	http.HandleFunc("/", transferRequestHandler)

	err := http.ListenAndServe(address, nil)
	if err != nil {
		fmt.Printf("Error occured")
	}

}
