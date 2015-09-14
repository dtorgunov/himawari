// himawari. A simple file upload handling server/client.
// Copyright (C) 2015 Denis Torgunov
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

// Package himawari implements a basic Himawari server and client.
package himawari

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

// defaultTimeout is the amount of time (in seconds), that the server
// will wait for a file upload after a request has been accepted.
const defaultTimeout = 60

// Pending is a wrapper around an array of PendingUpload that have been
// accepted, but haven't been uploaded nor timed out.
// It should be locked when in use, as the timing out routine runs in
// a separate thread.
type Pending struct {
	sync.Mutex
	uploads []PendingUpload
}

// writeError is a wrapper for reporting errors to the client.
// code should be a valid HTTP Response Code.
func writeError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	fmt.Fprintf(w, "%s\r\n", message)
}

// exists tests whether a file with a given filename exists.
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

// respond is used to construct a response to a request, either to
// accept it or to report an error.
func respond(w http.ResponseWriter, req Request, address string, directory string, readyQueue *Pending) error {
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
	resp := PendingUpload{Url: fmt.Sprintf("http://%s/%s", address, req.Filename), Filename: req.Filename, Timeout: defaultTimeout, Length: req.Length}
	json.NewEncoder(w).Encode(resp)
	readyQueue.uploads = append(readyQueue.uploads, resp)
	readyQueue.Unlock()

	return nil
}

// transferRequestHandler listens on the server root and accepts
// requests for uploads/file transfers.
func transferRequestHandler(w http.ResponseWriter, req *http.Request, address string, directory string, readyQueue *Pending) {
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

	err = respond(w, r, address, directory, readyQueue)
	if err != nil {
		log.Println(err)
	}

	fmt.Printf("%v\n", r)
	w.WriteHeader(204)
}

// removeRequest removes a given PendingUpload from a Pending queue. The
// caller MUST lock the queue before calling this function.
func removeRequest(u PendingUpload, readyQueue *Pending) {
	// assumes readyQueue is Locked by the caller
	var newQueue []PendingUpload
	for _, u1 := range readyQueue.uploads {
		if u != u1 {
			newQueue = append(newQueue, u1)
		} else {
			log.Printf("Removing %s from the queue", u.Filename)
		}
	}
	readyQueue.uploads = newQueue

}

// availableForUpload extracts a PendingUpload for a transfer of a given
// filename from a Pending queue. Returns an error if a request for
// such a transfer has not been accepted.
func availableForUpload(filename string, readyQueue *Pending) (PendingUpload, error) {
	readyQueue.Lock()
	defer readyQueue.Unlock()
	for _, u := range readyQueue.uploads {
		if u.Filename == filename {
			removeRequest(u, readyQueue)
			return u, nil
		}
	}
	return PendingUpload{}, errors.New("Unapproved upload")
}

// handleUpload saves the PUT payload to the corresponding location,
// given that the length of the file is as negotiated.
func handleUpload(w http.ResponseWriter, req *http.Request, expected PendingUpload, directory string) {
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

// requestHandlerConstructor constructs a Handler, closing around the
// address to listen on and a directory to save the uploads into.
func requestHandlerConstructor(address string, directory string, readyQueue *Pending) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/" { // root request
			transferRequestHandler(w, req, address, directory, readyQueue)
		} else {
			filename := req.URL.Path[1:]
			if r, err := availableForUpload(filename, readyQueue); err == nil {
				handleUpload(w, req, r, directory)
			} else {
				http.NotFound(w, req)
			}
		}
	}
}

// timeoutRequests checks a Pending queue for any requests that are
// past their timeout and removes them.
func timeoutRequests(readyQueue *Pending) {
	for {
		readyQueue.Lock()
		var newQueue []PendingUpload
		for _, u := range readyQueue.uploads {
			if u.Timeout > 0 {
				u.Timeout = u.Timeout - 10
				newQueue = append(newQueue, u)
			} else {
				log.Printf("Request for %s timed out", u.Filename)
			}
		}
		readyQueue.uploads = newQueue
		readyQueue.Unlock()
		time.Sleep(10 * time.Second)
	}
}

// StartServer is the main entry point for this package. It starts the
// server listening on the supplied address, and saving the uploaded
// files into the specified directory.
func StartServer(address string, datadir string) error {
	var readyQueue Pending

	http.HandleFunc("/", requestHandlerConstructor(address, datadir, &readyQueue))
	go timeoutRequests(&readyQueue)

	log.Printf("Starting a server on %s, serving the directory %s\n", address, datadir)
	err := http.ListenAndServe(address, nil)
	if err != nil {
		log.Printf("An error has occurred when starting the server")
		return err
	}

	return nil
}
