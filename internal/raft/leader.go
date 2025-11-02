package raft

import (
	"log"
	"sync"
	"time"

	"github.com/taylormeador/kv-store/internal/protocol"
)

func (n *Node) runHeartbeatLoop() {
	for {
		n.mu.RLock()
		state := n.state
		peers := n.peers
		n.mu.RUnlock()

		if state != LeaderState {
			return
		}

		for _, peer := range peers {
			go n.replicateToPeer(peer)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func (n *Node) replicateToPeer(peer string) (*AppendEntriesResponse, error) {
	n.mu.RLock()

	if n.state != LeaderState {
		n.mu.RUnlock()
		return nil, ErrNotLeader
	}

	nextIdx := n.nextIndex[peer]
	var entries []LogEntry
	if nextIdx <= len(n.log) {
		entries = n.log[nextIdx-1:]
	}

	var prevLogIndex, prevLogTerm int
	if nextIdx > 1 {
		prevLogEntry := n.log[nextIdx-2]
		prevLogIndex = prevLogEntry.Index
		prevLogTerm = prevLogEntry.Term
	}

	req := AppendEntriesRequest{
		Type:              AppendEntriesRPC,
		Term:              n.currentTerm,
		LeaderID:          n.id,
		PrevLogIndex:      prevLogIndex,
		PrevLogTerm:       prevLogTerm,
		Entries:           entries,
		LeaderCommitIndex: n.commitIndex,
	}
	n.mu.RUnlock()

	return n.sendAppendEntries(peer, req)
}

// Propose a new entry to the log and try to get quorum approval
func (n *Node) Propose(cmd protocol.Command) error {
	n.mu.Lock()
	if n.state != LeaderState {
		n.mu.Unlock()
		return ErrNotLeader
	}

	logEntry := LogEntry{
		Index:   n.getLastLogIndex() + 1,
		Term:    n.currentTerm,
		Command: cmd,
	}
	n.log = append(n.log, logEntry)
	n.storage.AppendEntry(logEntry)

	// Associate every entry with a channel.
	// Close the channel once the entry is replicated to a majority of nodes.
	commitCh := make(chan bool)
	n.commitWaiters[logEntry.Index] = commitCh

	n.mu.Unlock()

	var successMu sync.Mutex
	successes := 1
	successesNeeded := (len(n.peers)+1)/2 + 1
	for _, peer := range n.peers {
		go func(peer string) {
			log.Printf("replicating to peer %s", peer)
			resp, err := n.replicateToPeer(peer)
			if err != nil {
				log.Println(err)
				return
			}
			if resp.Success {
				n.mu.Lock()
				n.matchIndex[peer] = logEntry.Index
				n.nextIndex[peer] = logEntry.Index + 1
				n.mu.Unlock()

				successMu.Lock()
				successes++
				if successes >= successesNeeded {
					n.commitEntry(logEntry.Index)
				}
				successMu.Unlock()
			} else {
				// Follower has higher term, so we step down.
				if resp.Term > n.currentTerm {
					n.mu.Lock()
					n.becomeFollower(resp.Term)
					n.mu.Unlock()
					return
				}

				// Log is inconsistent, decrement until we catch them up
				n.mu.Lock()
				if n.nextIndex[peer] > 1 {
					n.nextIndex[peer]--
					log.Printf("follower %s rejected, backing up nextIndex to %d", peer, n.nextIndex[peer])
				}
				n.mu.Unlock()
			}
		}(peer)
	}

	select {
	case <-commitCh:
		return nil
	case <-time.After(5 * time.Second):
		return ErrTimeout
	}
}

// An entry is considered committed once the leader that created the entry has replicated it on a majority of the servers.
// Raft guarantees that commited entries are durable and will eventually be executed by all of the available state machines.
// This function closes the entry's channel once the majority of nodes have successfully replicated the entry.
func (n *Node) commitEntry(index int) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.commitIndex = index
	if ch, exists := n.commitWaiters[index]; exists {
		close(ch)
		delete(n.commitWaiters, index)
	}
}
