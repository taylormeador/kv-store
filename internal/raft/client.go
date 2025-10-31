package raft

import (
	"net"
)

// Send a request for vote to a peer
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

// AppendEntries RPC + heartbeat
func (n *Node) sendAppendEntries(peer string, req AppendEntriesRequest) (*AppendEntriesResponse, error) {
	var d net.Dialer
	conn, err := d.Dial("tcp", peer)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if err := n.writeJSON(conn, req); err != nil {
		return nil, err
	}

	var resp AppendEntriesResponse
	if err := n.readJSON(conn, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
