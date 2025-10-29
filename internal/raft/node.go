package raft

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
)

type NodeState string

const (
	LeaderState    NodeState = "LEADER"
	FollowerState  NodeState = "FOLLOWER"
	CandidateState NodeState = "CANDIDATE"
)

type Node struct {
	ID       int
	Port     int
	Peers    []int
	State    NodeState
	listener net.Listener
}

func NewNode(ID int, port int) *Node {
	return &Node{
		ID:    ID,
		Port:  port,
		State: FollowerState,
	}
}

func (n *Node) Listen() error {
	address := fmt.Sprintf(":%d", n.Port)
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	n.listener = ln

	log.Printf("Starting raft server on localhost:%d", n.Port)

	return nil
}

func (n *Node) Serve() {
	defer n.listener.Close()

	for {
		conn, err := n.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				break
			}
			log.Println(err)
			continue
		}
		log.Printf("Connected to %s", conn.RemoteAddr().String())
		go n.handleConnection(conn)
	}
}

func (n *Node) handleConnection(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()

		// Parse JSON message
		var msg map[string]any
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			log.Println("Parse error:", err)
			continue
		}

		// Route based on type
		msgType := msg["type"].(string)
		switch msgType {
		case "RequestVote":
			n.handleRequestVote(conn)
		case "AppendEntries":
			n.handleAppendEntries(conn)
		}

	}
}

func (n *Node) handleRequestVote(conn net.Conn) {
	response := RequestVoteResponse{
		Term:        1,
		VoteGranted: false,
	}

	jsonResponse, _ := json.Marshal(response)
	conn.Write(append(jsonResponse, '\n'))
}

func (n *Node) handleAppendEntries(conn net.Conn) {
	response := AppendEntriesResponse{
		Term:    1,
		Success: false,
	}

	jsonResponse, _ := json.Marshal(response)
	conn.Write(append(jsonResponse, '\n'))
}

func (n *Node) Shutdown() {
	log.Println("Shutting down raft node...")
	n.listener.Close()
}
