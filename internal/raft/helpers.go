package raft

import (
	"encoding/json"
	"net"
)

// Writes data as JSON for uniform responses
func (n *Node) writeJSON(conn net.Conn, data any) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	conn.Write(append(js, '\n'))

	return nil
}
