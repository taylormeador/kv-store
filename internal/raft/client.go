package raft

import (
	"net"
)

// Send RequestVote to a peer
func (n *Node) SendRequestVote() error {
	var d net.Dialer
	conn, err := d.Dial("tcp", n.Peers[0])
	if err != nil {
		return err
	}
	defer conn.Close()

	data := RequestVoteRequest{
		Type:         RequestVoteRPC,
		Term:         n.CurrentTerm,
		CandidateID:  n.ID,
		LastLogIndex: 1,
		LastLogTerm:  1,
	}
	err = n.writeJSON(conn, data)
	if err != nil {
		return err
	}

	return nil
}
