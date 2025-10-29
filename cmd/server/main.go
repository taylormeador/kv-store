package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/taylormeador/kv-store/internal/raft"
	"github.com/taylormeador/kv-store/internal/server"
)

var WAL_PATH = os.Getenv("WAL_PATH")

func main() {
	// Validate env vars
	KV_PORT, err := strconv.Atoi(os.Getenv("KV_PORT"))
	if err != nil {
		log.Fatal(err)
	}

	RAFT_PORT, err := strconv.Atoi(os.Getenv("RAFT_PORT"))
	if err != nil {
		log.Fatal(err)
	}

	RAFT_ID, err := strconv.Atoi(os.Getenv("RAFT_ID"))
	if err != nil {
		log.Fatal(err)
	}

	var WAL_PATH = os.Getenv("WAL_PATH")
	var RAFT_PEERS = os.Getenv("RAFT_PEERS")
	log.Println(RAFT_PEERS)

	// Set up interrupt channel
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// Set up raft
	raftNode := raft.NewNode(RAFT_ID, RAFT_PORT)
	err = raftNode.Listen()
	if err != nil {
		log.Fatal(err)
	}
	go raftNode.Serve()

	// Start kv server
	kvServer, err := server.NewServer(KV_PORT, WAL_PATH)
	if err != nil {
		log.Fatal(err)
	}

	err = kvServer.Listen()
	if err != nil {
		log.Fatal(err)
	}
	go kvServer.Serve()

	// Graceful shutdown
	signal := <-signals
	log.Println("Received signal:", signal)
	kvServer.Shutdown()
	raftNode.Shutdown()
	os.Exit(0)
}
