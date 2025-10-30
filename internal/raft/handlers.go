package raft

import (
	"log"
	"net"
	"time"
)

func (n *Node) handleRequestVote(conn net.Conn, req RequestVoteRequest) error {
	log.Printf("request vote received from %s", conn.RemoteAddr().String())

	n.mu.Lock()
	defer n.mu.Unlock()

	resp := RequestVoteResponse{
		Type:        RequestVoteRPC,
		Term:        n.currentTerm,
		VoteGranted: false,
	}

	// Reply false if term < currentTerm.
	if req.Term < n.currentTerm {
		log.Printf("candidate %d rejected for stale term", req.CandidateID)
		if err := n.writeJSON(conn, resp); err != nil {
			return err
		}
		return nil
	}

	if req.Term > n.currentTerm {
		log.Printf("higher term seen (%d), stepping down to follower", req.Term)
		n.becomeFollower(req.Term)
	}

	// If votedFor is null or candidateId, and candidate’s log is at
	// least as up-to-date as receiver’s log, grant vote.
	if n.votedFor == 0 || n.votedFor == req.CandidateID {
		if n.lastLogIndex <= req.LastLogIndex {
			resp.VoteGranted = true
			n.votedFor = req.CandidateID
			n.lastHeartbeat = time.Now()
			// TODO write currentTerm, votedFor, and log[] to disk before responding
		}
	}

	if err := n.writeJSON(conn, resp); err != nil {
		return err
	}
	log.Printf("vote %v", resp)
	return nil
}

func (n *Node) handleAppendEntries(conn net.Conn, req AppendEntriesRequest) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	resp := AppendEntriesResponse{
		Term:    n.currentTerm,
		Success: false,
	}

	if req.Term < n.currentTerm {
		if err := n.writeJSON(conn, resp); err != nil {
			return err
		}
		return nil
	} else {
		n.becomeFollower(req.Term)
		resp.Success = true
	}

	if err := n.writeJSON(conn, resp); err != nil {
		return err
	}
	return nil
}

func (n *Node) Shutdown() {
	log.Println("Shutting down raft node...")
	n.listener.Close()
}
