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
)

type Server struct {
	Port     int
	Store    *store.Store
	Raft     *raft.Node
	wg       sync.WaitGroup
	listener net.Listener
}

// Constructor
func NewServer(port int, wal_path string, raft *raft.Node) (*Server, error) {
	s := &Server{
		Port:  port,
		Store: store.NewStore(),
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

	log.Printf("starting server on localhost:%d", s.Port)

	return nil
}

// Serve() listens on the port and accepts new connections with a handler.
func (s *Server) Serve() {
	defer s.listener.Close()

	go s.consumeApplyCh()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				break
			}
			log.Println(err)
			continue
		}
		log.Printf("connected to %s", conn.RemoteAddr().String())
		s.wg.Add(1)
		go s.handleConnection(conn)
	}
}

func (s *Server) Shutdown() {
	// Close listener
	log.Println("shutting down server...")
	s.listener.Close()

	// Wait for workers to finish
	log.Println("listener closed, waiting for workers to finish...")
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	// Create a timeout for worker cleanup
	select {
	case <-done:
		log.Println("all workers exited...")
	case <-time.After(10 * time.Second):
		log.Println("worker cleanup timed out, forcing shutdown")
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
			log.Println("error parsing command:", err)

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
			if err := s.Raft.Propose(*c); err != nil {
				switch err {
				case raft.ErrNotLeader:
					// TODO set n.LeaderID somewhere
					response = fmt.Sprintf("%s - try %s", err.Error(), s.Raft.LeaderAddr)
				case raft.ErrTimeout:
					response = err.Error()
				default: // Should never get here
					response = err.Error()
				}

			} else {
				response = "OK"
			}
		case protocol.DeleteDirective:
			// Remember if exists before proposing delete
			exists := s.Store.Exists(c.Key)
			if exists {
				response = "TRUE"
			} else {
				response = "FALSE"
			}

			if err := s.Raft.Propose(*c); err != nil {
				response = err.Error()
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

// Apply the commands to the store in the background
func (s *Server) consumeApplyCh() {
	for entry := range s.Raft.ApplyCh {
		switch entry.Command.Directive {
		case protocol.SetDirective:
			log.Printf("applying %s", entry.Command.String())
			s.Store.Set(entry.Command.Key, entry.Command.Value)
		case protocol.DeleteDirective:
			log.Printf("applying %s", entry.Command.String())
			s.Store.Delete(entry.Command.Key)
		default:
			log.Printf("unknown entry in ApplyCh: %s", entry.Command.String())
		}
	}
}
