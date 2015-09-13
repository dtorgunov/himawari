package main

import "github.com/dtorgunov/himawari"

func main() {
	himawari.StartServer("localhost:3030", "data")
}
