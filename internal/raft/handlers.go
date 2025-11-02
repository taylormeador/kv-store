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
		if n.getLastLogIndex() <= req.LastLogIndex {
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
	log.Printf("receieved AppendEntries RPC")
	n.mu.Lock()
	defer n.mu.Unlock()

	resp := AppendEntriesResponse{
		Type:    AppendEntriesRPC,
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
	}

	// TODO: Check log consistency (prevLogIndex/prevLogTerm)
	// For now, just accept and append
	if len(req.Entries) > 0 {
		n.log = append(n.log, req.Entries...)
	}

	// Update commit index
	if req.LeaderCommitIndex > n.commitIndex {
		n.commitIndex = req.LeaderCommitIndex
	}

	resp.Success = true
	if err := n.writeJSON(conn, resp); err != nil {
		return err
	}
	return nil
}
