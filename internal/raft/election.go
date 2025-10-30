package raft

import (
	"log"
	"time"
)

// Periodically wakes up and checks if last heartbeat was within timeout duration
func (n *Node) runElectionTimer() {
	for {
		time.Sleep(50 * time.Millisecond)

		n.mu.RLock()
		state := n.state
		elapsed := time.Since(n.lastHeartbeat)
		timeout := n.heartbeatTimeout
		n.mu.RUnlock()
		if state != LeaderState {
			if elapsed > timeout {
				log.Println("Heartbeat timed out, starting election")
				n.startElection()
			}
		}
	}
}

// Become a Candidate and ask for votes to become leader
func (n *Node) startElection() {
	n.mu.Lock()
	n.state = CandidateState
	n.currentTerm++
	n.lastHeartbeat = time.Now()
	n.heartbeatTimeout = randomTimeout()

	// Remember what term this election started, since other vote requests
	// may mutate state in between vote request/response
	currentTerm := n.currentTerm

	// Vote for self and then wait for quorum
	n.votedFor = n.id
	votesReceived := 1
	votesNeeded := (len(n.peers)+1)/2 + 1

	// Release lock for RPC calls since they mutate state
	n.mu.Unlock()
	for _, peer := range n.peers {
		go func(peer string) {
			req := RequestVoteRequest{
				Type:        RequestVoteRPC,
				Term:        currentTerm,
				CandidateID: n.id,
			}
			resp, err := n.sendRequestVote(peer, req)
			if err != nil {
				log.Println(err)
				return
			}

			n.mu.Lock()
			defer n.mu.Unlock()

			// Check if we're still a Candidate and in the original term.
			// If not, someone became the leader already.
			if n.state != CandidateState || n.currentTerm != currentTerm {
				return
			}

			if resp.Term > n.currentTerm {
				log.Printf("%s has higher term %d, stepping down", peer, resp.Term)
				n.becomeFollower(resp.Term)
				return
			}

			if resp.VoteGranted {
				votesReceived++
				log.Printf("received vote from %s", peer)
				if votesReceived >= votesNeeded {
					log.Printf("election won for term %d", n.currentTerm)
					n.becomeLeader()
				}
			}
		}(peer)
	}
}

func (n *Node) becomeFollower(term int) {
	// Caller must hold lock
	n.state = FollowerState
	n.currentTerm = term
	n.votedFor = 0
	n.lastHeartbeat = time.Now()
}

func (n *Node) becomeLeader() {
	// Caller must hold lock
	n.state = LeaderState
	go n.runHeartbeatLoop()
}

func (n *Node) runHeartbeatLoop() {
	for {
		req := AppendEntriesRequest{
			Type:              AppendEntriesRPC,
			Term:              n.currentTerm,
			LeaderID:          n.id,
			PrevLogIndex:      0,
			PrevLogTerm:       0,
			Entries:           []string{""},
			LeaderCommitIndex: 0,
		}
		for _, peer := range n.peers {
			go n.sendAppendEntries(peer, req)
		}
		time.Sleep(50 * time.Millisecond)
	}
}
