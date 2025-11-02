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
			n.storage.SaveState(n.currentTerm, n.votedFor)
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
		Type:    AppendEntriesRPC,
		Term:    n.currentTerm,
		Success: false,
	}

	// Reply false if term < currentTerm
	if req.Term < n.currentTerm {
		if err := n.writeJSON(conn, resp); err != nil {
			return err
		}
		return nil
	} else {
		n.becomeFollower(req.Term)
	}

	// Reply false if log doesn't contain an entry at prevLogIndex
	// whose term matches prevLogTerm
	if req.PrevLogIndex > 0 {
		if req.PrevLogIndex > n.getLastLogIndex() {
			log.Printf("consistency check failed: missing entry at index %d (have up to %d)", req.PrevLogIndex, n.getLastLogIndex())
			if err := n.writeJSON(conn, resp); err != nil {
				return err
			}
			return nil
		}

		prevEntry := n.log[req.PrevLogIndex-1]
		if prevEntry.Term != req.PrevLogTerm {
			log.Printf("consistency check failed: term mismatch at index %d (have %d, need %d)", req.PrevLogIndex, prevEntry.Term, req.PrevLogTerm)
			log.Printf("deleting conflicting entries from index %d onwards", req.PrevLogIndex)
			n.log = n.log[:req.PrevLogIndex-1]
			if err := n.writeJSON(conn, resp); err != nil {
				return err
			}
			return nil
		}
	}

	// Last log entry is consistent so we append new entries,
	// checking for conflicts in existing entries and discarding
	insertIdx := req.PrevLogIndex + 1
	for i, entry := range req.Entries {
		logIdx := insertIdx + i
		if logIdx <= len(n.log) {
			existingEntry := n.log[logIdx-1]
			if existingEntry.Term != entry.Term {
				log.Printf("conflict at index %d: deleting from here", logIdx-1)
				n.log = n.log[:logIdx-1]
				n.log = append(n.log, req.Entries[i:]...)
				for _, newEntry := range req.Entries[i:] {
					n.storage.AppendEntry(newEntry)
				}
				break
			}
		} else {
			n.log = append(n.log, req.Entries[i:]...)
			for _, newEntry := range req.Entries[i:] {
				n.storage.AppendEntry(newEntry)
			}
			break
		}
	}

	// Update commit index
	if req.LeaderCommitIndex > n.commitIndex {
		n.commitIndex = min(req.LeaderCommitIndex, n.getLastLogIndex())
	}

	resp.Success = true
	if err := n.writeJSON(conn, resp); err != nil {
		return err
	}
	return nil
}
