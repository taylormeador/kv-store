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
	ID           int
	Port         int
	Peers        []string
	State        NodeState
	CurrentTerm  int
	VotedFor     int
	LastLogIndex int
	LastLogTerm  int
	Log          []string
	listener     net.Listener
}

func NewNode(ID int, port int, peers []string) *Node {
	return &Node{
		ID:    ID,
		Port:  port,
		Peers: peers,
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
		line := scanner.Bytes()

		// Parse JSON message
		var typeMsg struct {
			Type RPCType `json:"type"`
		}
		err := json.Unmarshal(line, &typeMsg)
		if err != nil {
			log.Println("Error parsing RPC type:", err)
			continue
		}

		// Route based on type
		switch typeMsg.Type {
		case RequestVoteRPC:
			var request RequestVoteRequest
			err = json.Unmarshal(line, &request)
			if err != nil {
				log.Println(err)
				continue
			}
			n.handleRequestVote(conn, request)
		case AppendEntriesRPC:
			n.handleAppendEntries(conn)
		default:
			log.Printf("Unknown RPC type: %s", typeMsg.Type)
		}
	}
}
