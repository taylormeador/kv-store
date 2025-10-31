package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
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

	RAFT_PEERS := strings.Split(os.Getenv("RAFT_PEERS"), ",")
	WAL_PATH = os.Getenv("WAL_PATH")

	// Set up interrupt channel
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// Set up raft
	raftNode := raft.NewNode(RAFT_ID, RAFT_PORT, RAFT_PEERS)
	err = raftNode.Listen()
	if err != nil {
		log.Fatal(err)
	}
	go raftNode.Start()

	// Start kv server
	kvServer, err := server.NewServer(KV_PORT, WAL_PATH, raftNode)
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
