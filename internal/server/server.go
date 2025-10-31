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
	"github.com/taylormeador/kv-store/internal/raft"
	"github.com/taylormeador/kv-store/internal/store"
	"github.com/taylormeador/kv-store/internal/wal"
)

type Server struct {
	Port     int
	Store    *store.Store
	WAL      *wal.WAL
	Raft     *raft.Node
	wg       sync.WaitGroup
	listener net.Listener
}

// Constructor
func NewServer(port int, wal_path string, raft *raft.Node) (*Server, error) {
	// Create WAL
	wal, err := wal.NewWAL(wal_path)
	if err != nil {
		return nil, err
	}

	// Init data store and restore to last known state
	store := store.NewStore()
	err = wal.Replay(store)
	if err != nil {
		return nil, err
	}

	s := &Server{
		Port:  port,
		Store: store,
		WAL:   wal,
		Raft:  raft,
	}
	return s, nil
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
	// Close listener
	log.Println("Shutting down server...")
	s.listener.Close()

	// Wait for workers to finish
	log.Println("Listener closed, waiting for workers to finish...")
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	// Create a timeout for worker cleanup
	select {
	case <-done:
		log.Println("All workers exited...")
	case <-time.After(10 * time.Second):
		log.Println("Worker cleanup timed out, forcing shutdown")
	}

	// Close WAL
	log.Println("Closing WAL...")
	if err := s.WAL.Close(); err != nil {
		log.Printf("Error closing WAL: %v", err)
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
		c, err := protocol.ParseCommand(ln) // TODO should this return value instead of pointer?
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
			s.WAL.Append(*c)
			if err := s.Raft.Propose(*c); err != nil {
				response = err.Error()
			} else {
				// TODO wait for raft to commit
				// TODO let Apply loop update store
				response = "OK"
			}
		case protocol.DeleteDirective:
			s.WAL.Append(*c)
			// TODO use raft here, need to implement apply loop first
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
		log.Printf("writing to %s: %s", conn.RemoteAddr().String(), response)
		_, err = conn.Write([]byte(response + "\n"))
		if err != nil {
			log.Println(err)
		}
	}

	log.Printf("closing connection with %s", conn.RemoteAddr().String())
}
