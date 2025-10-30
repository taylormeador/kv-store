package raft

import (
	"encoding/json"
	"log"
	"net"
)

func (n *Node) handleRequestVote(conn net.Conn, r RequestVoteRequest) error {
	log.Printf("Request vote received from %s", conn.RemoteAddr().String())

	var voteGranted bool

	// Reply false if term < currentTerm.
	if r.Term >= n.CurrentTerm {
		// If votedFor is null or candidateId, and candidate’s log is at
		// least as up-to-date as receiver’s log, grant vote.
		if n.VotedFor == 0 || n.VotedFor == r.CandidateID {
			if n.LastLogIndex <= r.LastLogIndex {
				voteGranted = true
			}
		}
	}

	response := RequestVoteResponse{
		Term:        n.CurrentTerm,
		VoteGranted: voteGranted,
	}
	err := n.writeJSON(conn, response)
	if err != nil {
		return err
	}

	log.Printf("Voted: %v", response)

	return nil
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
