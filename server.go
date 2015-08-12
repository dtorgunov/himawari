package main

import (
	"fmt"
	"net/http"
	"encoding/json"
)

type Request struct {
	Filename string
	Mime string
	Length int
}

// These configuration variables will be read from command line flags later
var (
	address = "localhost:3030"
	directory = "data"
)

func transferRequestHandler(w http.ResponseWriter, req *http.Request) {
	var r Request
	
	// Need to handle OPTIONS properly
	if (req.Method != "POST") {
		w.Header().Add("Allow", "POST")
		w.WriteHeader(405)
		return
	}

	// Need to check content-type
	err := json.NewDecoder(req.Body).Decode(&r)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, "The JSON data you have sent is invalid.\r\n")
		return
	}

	if r.Length <= 0 {
		w.WriteHeader(400)
		fmt.Fprintf(w, "Valid length needs to be specified\r\n")
		return
	}

	if r.Filename == "" && r.Mime == "" {
		w.WriteHeader(400)
		fmt.Fprintf(w, "Either a filename or MIME-type needs to be specified\r\n")
		return
	}

	fmt.Printf("%v\n", r)
	w.WriteHeader(204)
}

func main() {
	fmt.Printf("%s => %s\n", address, directory)
	http.HandleFunc("/", transferRequestHandler)

	err := http.ListenAndServe(address,nil)
	if err != nil {
		fmt.Printf("Error occured")
	}
	
}
