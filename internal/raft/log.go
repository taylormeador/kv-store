package raft

import "github.com/taylormeador/kv-store/internal/protocol"

type LogEntry struct {
	Index   int
	Term    int
	Command protocol.Command
}

func (n *Node) getLastLogIndex() int {
	if len(n.log) == 0 {
		return 0
	}
	return n.log[len(n.log)-1].Index
}

func (n *Node) getLastLogTerm() int {
	if len(n.log) == 0 {
		return 0
	}
	return n.log[len(n.log)-1].Term
}
