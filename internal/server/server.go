package server

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type Server struct {
	Port     int
	wg       sync.WaitGroup
	listener net.Listener
}

// Constructor
func NewServer(port int) *Server {
	return &Server{
		Port: port,
	}
}

// Opens listener.
func (s *Server) Listen() error {

	address := fmt.Sprintf(":%d", s.Port)
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	s.listener = ln

	log.Printf("Starting server on localhost:%d", s.Port)

	return nil
}

// Serve() listens on the port and accepts new connections with a handler.
func (s *Server) Serve() {
	defer s.listener.Close()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				break
			}
			log.Println(err)
			continue
		}
		log.Printf("Connected to %s", conn.RemoteAddr().String())
		s.wg.Add(1)
		go s.handleConnection(conn)
	}
}

func (s *Server) Shutdown() {
	// Close listener.
	log.Println("Shutting down server...")
	s.listener.Close()

	// Wait for workers to finish.
	log.Println("Listener closed, waiting for workers to finish...")
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	// Create a timeout for worker cleanup.
	select {
	case <-done:
		log.Println("All workers exited, stopping now")
	case <-time.After(10 * time.Second):
		log.Println("Worker cleanup timed out, forcing shutdown")
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	defer s.wg.Done()

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
