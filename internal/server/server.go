package server

import (
	"fmt"
	"log"
	"net"
)

type Server struct {
	Port int
}

// Start() listens on the port and accepts new connections with a handler.
func (s *Server) Start() error {
	address := fmt.Sprintf(":%d", s.Port)
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		log.Printf("Connected to %s", conn.RemoteAddr().String())
		if err != nil {
			log.Println(err)
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Enter read/respond loop
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			log.Println(err)
			break
		}
		message := string(buf[:n])
		log.Println(message)

		// Send OK response
		_, err = conn.Write([]byte("OK\n"))
		if err != nil {
			log.Println(err)
		}
	}

	log.Printf("Closing connection with %s", conn.RemoteAddr().String())
}
