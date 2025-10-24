package wal

import (
	"bufio"
	"os"

	"github.com/taylormeador/kv-store/internal/protocol"
	"github.com/taylormeador/kv-store/internal/store"
)

type WAL struct {
	filepath string
	file     *os.File
}

// Creates a WAL at the given filepath
func NewWAL(filepath string) (*WAL, error) {
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	return &WAL{file: file, filepath: filepath}, err
}

// Close the file
func (w *WAL) Close() error {
	return w.file.Close()
}

// Write command to file and fsync
func (w *WAL) Append(command protocol.Command) error {
	_, err := w.file.WriteString(command.String() + "\n")
	if err != nil {
		return err
	}
	return w.file.Sync()
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
