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

package main

import (
	"flag"
	"fmt"
	"github.com/dtorgunov/himawari"
	"os"
)

var (
	filename = flag.String("f", "", "the file to upload")
	server   = flag.String("s", "", "the server to upload it to")
)

func main() {
	flag.Parse()
	if (len(*filename) < 1) || (len(*server) < 1) {
		flag.Usage()
		os.Exit(1)
	}
	s := himawari.SendFile(*filename, *server)
	fmt.Println(s)
}
