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

package himawari

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

// negotiateUpload sends a request to upload a file to the server and
// returns the response (a PendingUpload) or errors, if any
func negotiateUpload(filename string, serverUrl string) (PendingUpload, error) {
	log.Printf("Preparing to send file %s to %s\n", filename, serverUrl)

	file, err := os.Open(filename)
	if err != nil {
		e := fmt.Sprintf("An error occurred while opening file to determine its length: %s", err)
		return PendingUpload{}, errors.New(e)
	}
	defer file.Close()

	fs, err := file.Stat()
	if err != nil {
		e := fmt.Sprintf("An error occurred while trying to stat the file: %s", err)
		return PendingUpload{}, errors.New(e)
	}

	req := Request{Filename: filename, Length: int(fs.Size())}
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(req)

	reqStr := buf.String()
	resp, err := http.Post(serverUrl, "text/json", strings.NewReader(reqStr))
	if err != nil {
		e := fmt.Sprintf("Error while sending request: %s", err)
		return PendingUpload{}, errors.New(e)
	}
	defer resp.Body.Close()

	var response PendingUpload
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		e := fmt.Sprintf("Error while parsing response: %s", err)
		return PendingUpload{}, errors.New(e)
	}

	return response, nil
}

// upload sends the file to the server, once the request for transfer
// has been accepted. It returns the URL at which the uploaded file
// can be reached, and an error, if any.
func upload(filename string, u PendingUpload) (string, error) {
	log.Printf("Sending file %s", filename)

	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		e := fmt.Sprintf("An error occurred while reading file: %s", err)
		return "", errors.New(e)
	}

	req, err := http.NewRequest("PUT", u.Url, bytes.NewReader(buf))
	if err != nil {
		e := fmt.Sprintf("Error constructing request: %s", err)
		return "", errors.New(e)
	}

	var client http.Client
	resp, err := client.Do(req)
	if err != nil {
		e := fmt.Sprintf("An error occured while sending request: %s", err)
		return "", errors.New(e)
	}
	defer resp.Body.Close()

	// Protocol errors are treated in a generic manner for now
	if resp.StatusCode != 201 {
		errorMsg, _ := ioutil.ReadAll(resp.Body)
		log.Fatalf("A protocol error has occurred (%d): %s", resp.StatusCode, errorMsg)
	}

	return resp.Header.Get("Location"), nil
}

// SendFile is used to upload a file to the server. It returns a
// string if successful and terminates the client if an error occurs.
func SendFile(filename string, server string) string {
	u, err := negotiateUpload(filename, server)
	if err != nil {
		log.Fatal(err)
	}

	url, err := upload(filename, u)
	if err != nil {
		log.Fatal(err)
	}

	return url
}
