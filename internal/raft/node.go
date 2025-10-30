package raft

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type NodeState string

const (
	LeaderState    NodeState = "LEADER"
	FollowerState  NodeState = "FOLLOWER"
	CandidateState NodeState = "CANDIDATE"
)

type Node struct {
	// Self
	id    int
	port  int
	peers []string

	// State
	mu           sync.RWMutex
	state        NodeState
	currentTerm  int
	votedFor     int
	lastLogIndex int
	lastLogTerm  int
	log          []string

	// Timing
	lastHeartbeat    time.Time
	heartbeatTimeout time.Duration

	listener net.Listener
}

func NewNode(ID int, port int, peers []string) *Node {
	return &Node{
		id:    ID,
		port:  port,
		peers: peers,

		state: FollowerState,

		lastHeartbeat:    time.Now(),
		heartbeatTimeout: randomTimeout(),
	}
}

func (n *Node) Listen() error {
	address := fmt.Sprintf(":%d", n.port)
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	n.listener = ln

	log.Printf("Starting raft server on localhost:%d", n.port)

	return nil
}

func (n *Node) Start() {
	defer n.listener.Close()

	go n.runElectionTimer()

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
			var req RequestVoteRequest
			err = json.Unmarshal(line, &req)
			if err != nil {
				log.Println(err)
				continue
			}
			n.handleRequestVote(conn, req)
		case AppendEntriesRPC:
			var req AppendEntriesRequest
			err = json.Unmarshal(line, &req)
			if err != nil {
				log.Println(err)
				continue
			}
			n.handleAppendEntries(conn, req)
		default:
			log.Printf("Unknown RPC type: %s", typeMsg.Type)
		}
	}
}
