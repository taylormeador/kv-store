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
	// Set up interrupt channel
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// Start server
	s := server.NewServer(PORT_NUMBER)
	err := s.Listen()
	if err != nil {
		log.Fatal("Problem opening listener", err)
	}
	go s.Serve()

	// Graceful shutdown
	signal := <-signals
	log.Println("Received signal:", signal)
	s.Shutdown()
	os.Exit(0)
}
