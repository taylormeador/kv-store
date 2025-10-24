package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/taylormeador/kv-store/internal/server"
)

const PORT_NUMBER int = 8080
const WAL_PATH string = "/var/lib/kv-store/WAL.log"

func main() {
	// Set up interrupt channel
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// Start server
	s, err := server.NewServer(PORT_NUMBER, WAL_PATH)
	if err != nil {
		log.Fatal(err)
	}

	err = s.Listen()
	if err != nil {
		log.Fatal(err)
	}
	go s.Serve()

	// Graceful shutdown
	signal := <-signals
	log.Println("Received signal:", signal)
	s.Shutdown()
	os.Exit(0)
}
