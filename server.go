package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Request struct {
	Filename string
	Mime     string
	Length   int
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

	if r.Length <= 0 {
		writeError(w, 400, "Valid length needs to be specified.")
		return
	}

	if r.Filename == "" && r.Mime == "" {
		writeError(w, 400, "Either a filename or MIME-type needs to be specified")
		return
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
