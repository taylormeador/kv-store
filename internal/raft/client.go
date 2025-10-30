package raft

import (
	"net"
)

// Send a request to vote to a peer
func (n *Node) sendRequestVote(peer string, req RequestVoteRequest) (*RequestVoteResponse, error) {
	var d net.Dialer
	conn, err := d.Dial("tcp", peer)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if err := n.writeJSON(conn, req); err != nil {
		return nil, err
	}

	var resp RequestVoteResponse
	if err := n.readJSON(conn, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
