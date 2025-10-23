package server

import (
	"fmt"
	"log"
	"net"
)

type Server struct {
	Port     int
	numConns int
}

// Start() listens on the port and accepts new connections with a handler.
func (s *Server) Start() {
	address := fmt.Sprintf(":%d", s.Port)
	ln, err := net.Listen("tcp", address)
	if err != nil {
		log.Println(err)
		return
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		log.Printf("Connected to %s", conn.RemoteAddr().String())
		s.numConns += 1
		log.Printf("%d connections now open", s.numConns)
		if err != nil {
			log.Println(err)
			return
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
		_, err = conn.Write([]byte("OK\r\n"))
		if err != nil {
			log.Println(err)
		}
	}
	s.numConns -= 1
	log.Printf("%d connnections now open", s.numConns)
}
