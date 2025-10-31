package raft

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"time"
)

var ErrNotLeader = errors.New("ERROR not leader")
var ErrTimeout = errors.New("ERROR timeout")

// Reads RPCs as JSON
func (n *Node) readJSON(conn net.Conn, target any) error {
	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() {
		return fmt.Errorf("no data")
	}
	return json.Unmarshal(scanner.Bytes(), target)
}

// Writes data as JSON for uniform responses
func (n *Node) writeJSON(conn net.Conn, data any) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	conn.Write(append(js, '\n'))

	return nil
}

// Returns a random timeout between 150-300 ms
func randomTimeout() time.Duration {
	min := 150
	max := 300
	ms := rand.Intn(max-min) + min
	return time.Duration(ms) * time.Millisecond
}
