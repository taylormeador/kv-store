package wal

import (
	"bufio"
	"os"
	"sync"
	"time"

	"github.com/taylormeador/kv-store/internal/protocol"
	"github.com/taylormeador/kv-store/internal/store"
)

type WAL struct {
	filepath   string
	file       *os.File
	mu         sync.Mutex
	shutdownCh chan struct{}
	wg         sync.WaitGroup
}

// Creates a WAL at the given filepath
func NewWAL(filepath string) (*WAL, error) {
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	w := &WAL{
		filepath:   filepath,
		file:       file,
		shutdownCh: make(chan struct{}),
	}

	// Start background fsync goroutine (every 1 second)
	w.wg.Add(1)
	go w.fsyncLoop()

	return w, nil
}

// Close the file
func (w *WAL) Close() error {
	close(w.shutdownCh)

	// Wait for final fsync
	w.wg.Wait()

	return w.file.Close()
}

// Write command to file
func (w *WAL) Append(command protocol.Command) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Write to OS buffer, let fsyncLoop() handle flush
	_, err := w.file.WriteString(command.String() + "\n")
	if err != nil {
		return err
	}
	return nil
}

// Background goroutine that fsyncs every second
func (w *WAL) fsyncLoop() {
	defer w.wg.Done()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.mu.Lock()
			w.file.Sync() // Force buffered data to disk
			w.mu.Unlock()

		case <-w.shutdownCh:
			// Final fsync on shutdown
			w.mu.Lock()
			w.file.Sync()
			w.mu.Unlock()
			return
		}
	}
}

// Replays WAL and executes on store, restoring last known state
func (w *WAL) Replay(s *store.Store) error {
	readFile, err := os.Open(w.filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No WAL file yet
		}
		return err
	}
	defer readFile.Close()

	scanner := bufio.NewScanner(readFile)
	for scanner.Scan() {
		ln := scanner.Text()
		c, err := protocol.ParseCommand(ln)
		if err != nil {
			return err
		}

		// Execute only GET/SET commands
		switch c.Directive {
		case protocol.SetDirective:
			s.Set(c.Key, c.Value)
		case protocol.DeleteDirective:
			s.Delete(c.Key)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}
