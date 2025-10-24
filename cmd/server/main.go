package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/taylormeador/kv-store/internal/server"
)

const PORT_NUMBER int = 8080

func main() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	s := &server.Server{
		Port: PORT_NUMBER,
	}
	log.Printf("Starting server on localhost:%d", PORT_NUMBER)
	go func() {
		if err := s.Start(); err != nil {
			log.Fatal("Failed to start server:", err)
		}
	}()

	signal := <-signals
	log.Println("Received signal:", signal)
	s.Shutdown()
	os.Exit(0)
}
