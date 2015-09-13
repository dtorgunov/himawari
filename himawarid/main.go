// himawari. A simple file upload handling server.
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

// This package is used to start a basic himawari server
package main

import (
	"flag"
	"github.com/dtorgunov/himawari"
)

var (
	address = flag.String("address", ":3030", "the address to listen on")
	datadir = flag.String("datadir", "data", "the directory to save the uploaded files to")
)

func main() {
	flag.Parse()
	himawari.StartServer(*address, *datadir)
}
