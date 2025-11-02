package raft

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"

	"github.com/taylormeador/kv-store/internal/protocol"
)

type Storage struct {
	filepath string
	file     *os.File
	mu       sync.Mutex
}

func NewStorage(filepath string) (*Storage, error) {
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return &Storage{
		filepath: filepath,
		file:     file,
	}, nil
}

func (s *Storage) Close() error {
	return s.file.Close()
}

type PersistentState struct {
	Type     string `json:"type"` // "state"
	Term     int    `json:"term"`
	VotedFor int    `json:"votedFor"`
}

type PersistentEntry struct {
	Type    string           `json:"type"` // "entry"
	Index   int              `json:"index"`
	Term    int              `json:"term"`
	Command protocol.Command `json:"command"`
}

func (s *Storage) SaveState(term int, votedFor int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state := PersistentState{
		Type:     "state",
		Term:     term,
		VotedFor: votedFor,
	}

	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	_, err = s.file.Write(append(data, '\n'))
	if err != nil {
		return err
	}

	return s.file.Sync()
}

// Append a log entry
func (s *Storage) AppendEntry(entry LogEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	persistEntry := PersistentEntry{
		Type:    "entry",
		Index:   entry.Index,
		Term:    entry.Term,
		Command: entry.Command,
	}

	data, err := json.Marshal(persistEntry)
	if err != nil {
		return err
	}

	_, err = s.file.Write(append(data, '\n'))
	if err != nil {
		return err
	}

	return s.file.Sync()
}

// Truncate log (when deleting conflicting entries)
func (s *Storage) TruncateLog(fromIndex int) error {
	// This is complex - for now, rewrite entire file
	// Later optimization: use separate log file with truncate()
	return nil // TODO: implement if needed
}

func (s *Storage) Restore() (int, int, []LogEntry, error) {
	file, err := os.Open(s.filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, 0, []LogEntry{}, nil
		}
		return 0, 0, nil, err
	}
	defer file.Close()

	var currentTerm, votedFor int
	var log []LogEntry

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Bytes()

		var typeCheck struct {
			Type string `json:"type"`
		}
		json.Unmarshal(line, &typeCheck)

		switch typeCheck.Type {
		case "state":
			var state PersistentState
			if err := json.Unmarshal(line, &state); err != nil {
				return 0, 0, nil, err
			}
			currentTerm = state.Term
			votedFor = state.VotedFor

		case "entry":
			var entry PersistentEntry
			if err := json.Unmarshal(line, &entry); err != nil {
				return 0, 0, nil, err
			}
			log = append(log, LogEntry{
				Index:   entry.Index,
				Term:    entry.Term,
				Command: entry.Command,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, 0, nil, err
	}

	return currentTerm, votedFor, log, nil
}
