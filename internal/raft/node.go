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
	id         int
	port       int
	listener   net.Listener
	peers      []string
	LeaderAddr string // TODO implement this

	// State
	mu          sync.RWMutex
	state       NodeState
	currentTerm int
	votedFor    int
	log         []LogEntry // LogEntry.Index is 1-indexed while log is 0-indexed
	ApplyCh     chan LogEntry
	storage     *Storage

	// Volatile state
	commitIndex   int
	lastApplied   int
	commitWaiters map[int]chan bool

	// Leader volatile state
	nextIndex  map[string]int // peer -> index of next log entry to send to node
	matchIndex map[string]int // peer -> index of highest log entry known to be replicated on node

	// Timing
	lastHeartbeat    time.Time
	heartbeatTimeout time.Duration
}

func NewNode(ID int, port int, peers []string, storagePath string) (*Node, error) {
	storage, err := NewStorage(storagePath)
	if err != nil {
		return nil, err
	}

	term, votedFor, logEntries, err := storage.Restore()
	if err != nil {
		return nil, err
	}

	n := &Node{
		id:               ID,
		port:             port,
		peers:            peers,
		state:            FollowerState,
		currentTerm:      term,
		votedFor:         votedFor,
		log:              logEntries,
		ApplyCh:          make(chan LogEntry, 100),
		storage:          storage,
		commitWaiters:    make(map[int]chan bool),
		lastHeartbeat:    time.Now(),
		heartbeatTimeout: randomTimeout(),
	}
	log.Printf("restored from disk: term=%d, votedFor:=%d, log entries=%d", term, votedFor, len(logEntries))
	return n, nil
}

func (n *Node) Shutdown() {
	log.Println("shutting down raft node...")
	n.listener.Close()
}

func (n *Node) Listen() error {
	address := fmt.Sprintf(":%d", n.port)
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	n.listener = ln

	log.Printf("starting raft server on localhost:%d", n.port)

	return nil
}

func (n *Node) Start() {
	defer n.listener.Close()

	go n.runElectionTimer()
	go n.runApplyLoop()

	for {
		conn, err := n.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				break
			}
			log.Println(err)
			continue
		}
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
			log.Println("error parsing RPC type:", err)
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
			log.Printf("unknown RPC type: %s", typeMsg.Type)
		}
	}
}

func (n *Node) runApplyLoop() {
	for {
		time.Sleep(10 * time.Millisecond)

		n.mu.Lock()
		for n.commitIndex > n.lastApplied {
			n.lastApplied++
			if n.lastApplied > len(n.log) {
				break
			} // TODO is this actually necessary?

			log.Printf("sending %v on ApplyCh", n.log[n.lastApplied-1])
			n.ApplyCh <- n.log[n.lastApplied-1]
		}
		n.mu.Unlock()
	}
}
