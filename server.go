package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// A minute
const defaultTimeout = 60

type Request struct {
	Filename string
	Mime     string
	Length   int
}

type Response struct {
	Url      string `json:"url"`
	Timeout  int    `json:"timeout"`
	Filename string `json:"filename"`
	Length   int    `json:"-"`
}

type Pending struct {
	sync.Mutex
	responses []Response
}

// These configuration variables will be read from command line flags later
var (
	address   = "localhost:3030"
	directory = "data"
)

var readyQueue Pending

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

	readyQueue.Lock()
	resp := Response{Url: fmt.Sprintf("http://%s/%s", address, req.Filename), Filename: req.Filename, Timeout: defaultTimeout, Length: req.Length}
	json.NewEncoder(w).Encode(resp)
	readyQueue.responses = append(readyQueue.responses, resp)
	readyQueue.Unlock()

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

func removeRequest(r Response) {
	// assumes readyQueue is Locked by the caller
	var newQueue []Response
	for _, r1 := range readyQueue.responses {
		if r != r1 {
			newQueue = append(newQueue, r1)
		} else {
			log.Printf("Removing %s from the queue", r.Filename)
		}
	}
	readyQueue.responses = newQueue

}

func availableForUpload(filename string) (Response, error) {
	readyQueue.Lock()
	defer readyQueue.Unlock()
	for _, r := range readyQueue.responses {
		if r.Filename == filename {
			removeRequest(r)
			return r, nil
		}
	}
	return Response{}, errors.New("Unapproved upload")
}

func handleUpload(w http.ResponseWriter, req *http.Request, expected Response) {
	if req.Method != "PUT" {
		w.Header().Add("Allow", "PUT")
		w.WriteHeader(405)
		return
	}

	if int64(expected.Length) != req.ContentLength {
		writeError(w, 400, "The length of the content doesn't match the originally negotiated length")
		log.Printf("Expected %v bytes, received %v bytes. Upload aborted.\n", expected.Length, req.ContentLength)
		return
	}

	file, err := os.Create(fmt.Sprintf("%s/%s", directory, expected.Filename))
	if err != nil {
		log.Printf("Error creating file: %v\n", err)
		writeError(w, 500, "Could not create resource")
		return
	}
	defer file.Close()

	var buf []byte
	buf, err = ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error when reading request body: %v\n", err)
		writeError(w, 500, "Could not read the content")
		return
	}

	_, err = file.Write(buf)
	if err != nil {
		log.Printf("Error when writing to file: %v\n", err)
		writeError(w, 500, "Could not write file")
		return
	}

	w.Header().Add("Location", expected.Url)
	w.WriteHeader(201)
	fmt.Fprintf(w, "%s", expected.Url)
}

func requestHandler(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/" { // root request
		transferRequestHandler(w, req)
	} else {
		filename := req.URL.Path[1:]
		if r, err := availableForUpload(filename); err == nil {
			handleUpload(w, req, r)
		} else {
			http.NotFound(w, req)
		}
	}
}

func timeoutRequests() {
	for {
		readyQueue.Lock()
		var newQueue []Response
		for _, r := range readyQueue.responses {
			if r.Timeout > 0 {
				r.Timeout = r.Timeout - 10
				newQueue = append(newQueue, r)
			} else {
				log.Printf("Request for %s timed out", r.Filename)
			}
		}
		readyQueue.responses = newQueue
		readyQueue.Unlock()
		time.Sleep(10 * time.Second)
	}
}

func main() {
	fmt.Printf("%s => %s\n", address, directory)

	http.HandleFunc("/", requestHandler)

	go timeoutRequests()

	err := http.ListenAndServe(address, nil)
	if err != nil {
		fmt.Printf("Error occured")
	}

}
