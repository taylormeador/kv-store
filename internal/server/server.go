package server

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/taylormeador/kv-store/internal/protocol"
	"github.com/taylormeador/kv-store/internal/store"
)

type Server struct {
	Port     int
	Store    *store.Store
	wg       sync.WaitGroup
	listener net.Listener
}

// Constructor
func NewServer(port int) *Server {
	return &Server{
		Port:  port,
		Store: store.NewStore(),
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
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		// Parse command
		ln := scanner.Text()
		c, err := protocol.ParseCommand(ln)
		if err != nil {
			log.Println("Error parsing command:", err)

			// Send response
			response := fmt.Sprintf("ERROR %s\n", err)
			_, err = conn.Write([]byte(response))
			if err != nil {
				log.Println(err)
			}
			continue
		}

		// Execute command
		response := ""
		switch c.Directive {
		case protocol.GetDirective:
			val, exists := s.Store.Get(c.Key)
			if !exists {
				response = "ERROR key not found"
			} else {
				response = val
			}
		case protocol.SetDirective:
			s.Store.Set(c.Key, c.Value)
			response = "OK"
		case protocol.DeleteDirective:
			exists := s.Store.Delete(c.Key)
			if exists {
				response = "TRUE"
			} else {
				response = "FALSE"
			}
		case protocol.ExistsDirective:
			exists := s.Store.Exists(c.Key)
			if exists {
				response = "TRUE"
			} else {
				response = "FALSE"
			}
		}

		// Send response
		log.Printf("Writing to %s: %s", conn.RemoteAddr().String(), response)
		_, err = conn.Write([]byte(response + "\n"))
		if err != nil {
			log.Println(err)
		}
	}

	log.Printf("Closing connection with %s", conn.RemoteAddr().String())
}
