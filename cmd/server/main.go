package main

import (
	"log"

	"github.com/taylormeador/kv-store/internal/server"
)

const PORT_NUMBER int = 8080

func main() {
	s := &server.Server{
		Port: PORT_NUMBER,
	}
	log.Printf("Starting server on localhost:%d", PORT_NUMBER)
	s.Start()
}
