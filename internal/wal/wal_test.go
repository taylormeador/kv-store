package wal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/taylormeador/kv-store/internal/protocol"
	"github.com/taylormeador/kv-store/internal/store"
)

func TestNewWAL(t *testing.T) {
	tempDir := t.TempDir()
	testPath := filepath.Join(tempDir, "test.log")

	wal, err := NewWAL(testPath)
	if err != nil {
		t.Fatalf("NewWAL() error = %v", err)
	}
	defer wal.Close()

	if wal.file == nil {
		t.Error("NewWAL() did not initialize file handle")
	}

	// Check file was created
	if _, err := os.Stat(testPath); os.IsNotExist(err) {
		t.Error("NewWAL() did not create file")
	}
}

func TestAppend(t *testing.T) {
	tempDir := t.TempDir()
	testPath := filepath.Join(tempDir, "test.log")

	wal, err := NewWAL(testPath)
	if err != nil {
		t.Fatalf("NewWAL() error = %v", err)
	}
	defer wal.Close()

	cmd := protocol.Command{
		Directive: protocol.SetDirective,
		Key:       "foo",
		Value:     "bar",
	}

	err = wal.Append(cmd)
	if err != nil {
		t.Errorf("Append() error = %v", err)
	}

	// Read file and verify content
	content, err := os.ReadFile(testPath)
	if err != nil {
		t.Fatalf("Failed to read WAL file: %v", err)
	}

	expected := "SET foo bar\n"
	if string(content) != expected {
		t.Errorf("WAL content = %q, want %q", string(content), expected)
	}
}

func TestAppendMultiple(t *testing.T) {
	tempDir := t.TempDir()
	testPath := filepath.Join(tempDir, "test.log")

	wal, err := NewWAL(testPath)
	if err != nil {
		t.Fatalf("NewWAL() error = %v", err)
	}
	defer wal.Close()

	commands := []protocol.Command{
		{Directive: protocol.SetDirective, Key: "key1", Value: "value1"},
		{Directive: protocol.SetDirective, Key: "key2", Value: "value2"},
		{Directive: protocol.DeleteDirective, Key: "key1", Value: ""},
	}

	for _, cmd := range commands {
		if err := wal.Append(cmd); err != nil {
			t.Errorf("Append() error = %v", err)
		}
	}

	content, err := os.ReadFile(testPath)
	if err != nil {
		t.Fatalf("Failed to read WAL file: %v", err)
	}

	expected := "SET key1 value1\nSET key2 value2\nDELETE key1\n"
	if string(content) != expected {
		t.Errorf("WAL content = %q, want %q", string(content), expected)
	}
}

func TestReplayEmpty(t *testing.T) {
	tempDir := t.TempDir()
	testPath := filepath.Join(tempDir, "test.log")

	// Create empty WAL file
	file, err := os.Create(testPath)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file.Close()

	wal, err := NewWAL(testPath)
	if err != nil {
		t.Fatalf("NewWAL() error = %v", err)
	}
	defer wal.Close()

	s := store.NewStore()
	err = wal.Replay(s)
	if err != nil {
		t.Errorf("Replay() error = %v", err)
	}
}

func TestReplaySingleCommand(t *testing.T) {
	tempDir := t.TempDir()
	testPath := filepath.Join(tempDir, "test.log")

	// Write a command to the file
	file, err := os.Create(testPath)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file.WriteString("SET foo bar\n")
	file.Close()

	wal, err := NewWAL(testPath)
	if err != nil {
		t.Fatalf("NewWAL() error = %v", err)
	}
	defer wal.Close()

	s := store.NewStore()
	err = wal.Replay(s)
	if err != nil {
		t.Errorf("Replay() error = %v", err)
	}

	// Verify the store has the key
	value, exists := s.Get("foo")
	if !exists {
		t.Error("Replay() did not restore key 'foo'")
	}
	if value != "bar" {
		t.Errorf("Replay() restored value = %q, want %q", value, "bar")
	}
}

func TestReplayMultipleCommands(t *testing.T) {
	tempDir := t.TempDir()
	testPath := filepath.Join(tempDir, "test.log")

	// Write multiple commands
	file, err := os.Create(testPath)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file.WriteString("SET key1 value1\n")
	file.WriteString("SET key2 value2\n")
	file.WriteString("SET key3 value3\n")
	file.Close()

	wal, err := NewWAL(testPath)
	if err != nil {
		t.Fatalf("NewWAL() error = %v", err)
	}
	defer wal.Close()

	s := store.NewStore()
	err = wal.Replay(s)
	if err != nil {
		t.Errorf("Replay() error = %v", err)
	}

	// Verify all keys
	expected := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	for key, want := range expected {
		got, exists := s.Get(key)
		if !exists {
			t.Errorf("Replay() did not restore key %q", key)
		}
		if got != want {
			t.Errorf("Replay() restored %q = %q, want %q", key, got, want)
		}
	}
}

func TestReplayWithOverwrite(t *testing.T) {
	tempDir := t.TempDir()
	testPath := filepath.Join(tempDir, "test.log")

	// Write commands that overwrite same key
	file, err := os.Create(testPath)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file.WriteString("SET foo value1\n")
	file.WriteString("SET foo value2\n")
	file.WriteString("SET foo value3\n")
	file.Close()

	wal, err := NewWAL(testPath)
	if err != nil {
		t.Fatalf("NewWAL() error = %v", err)
	}
	defer wal.Close()

	s := store.NewStore()
	err = wal.Replay(s)
	if err != nil {
		t.Errorf("Replay() error = %v", err)
	}

	// Should have the last value
	value, exists := s.Get("foo")
	if !exists {
		t.Error("Replay() did not restore key 'foo'")
	}
	if value != "value3" {
		t.Errorf("Replay() restored value = %q, want %q", value, "value3")
	}
}

func TestReplayWithDelete(t *testing.T) {
	tempDir := t.TempDir()
	testPath := filepath.Join(tempDir, "test.log")

	// Write SET then DELETE
	file, err := os.Create(testPath)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file.WriteString("SET foo bar\n")
	file.WriteString("DELETE foo\n")
	file.Close()

	wal, err := NewWAL(testPath)
	if err != nil {
		t.Fatalf("NewWAL() error = %v", err)
	}
	defer wal.Close()

	s := store.NewStore()
	err = wal.Replay(s)
	if err != nil {
		t.Errorf("Replay() error = %v", err)
	}

	// Key should not exist
	_, exists := s.Get("foo")
	if exists {
		t.Error("Replay() key 'foo' should not exist after DELETE")
	}
}

func TestReplayIgnoresGetAndExists(t *testing.T) {
	tempDir := t.TempDir()
	testPath := filepath.Join(tempDir, "test.log")

	// Write GET and EXISTS commands (should be ignored)
	file, err := os.Create(testPath)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file.WriteString("GET foo\n")
	file.WriteString("EXISTS bar\n")
	file.WriteString("SET baz qux\n")
	file.Close()

	wal, err := NewWAL(testPath)
	if err != nil {
		t.Fatalf("NewWAL() error = %v", err)
	}
	defer wal.Close()

	s := store.NewStore()
	err = wal.Replay(s)
	if err != nil {
		t.Errorf("Replay() error = %v", err)
	}

	// Only SET should have been executed
	if _, exists := s.Get("foo"); exists {
		t.Error("Replay() should not have created 'foo' from GET command")
	}
	if _, exists := s.Get("bar"); exists {
		t.Error("Replay() should not have created 'bar' from EXISTS command")
	}
	if _, exists := s.Get("baz"); !exists {
		t.Error("Replay() should have created 'baz' from SET command")
	}
}

func TestAppendAndReplay(t *testing.T) {
	tempDir := t.TempDir()
	testPath := filepath.Join(tempDir, "test.log")

	// Write some commands
	wal1, err := NewWAL(testPath)
	if err != nil {
		t.Fatalf("NewWAL() error = %v", err)
	}

	commands := []protocol.Command{
		{Directive: protocol.SetDirective, Key: "user:1", Value: "alice"},
		{Directive: protocol.SetDirective, Key: "user:2", Value: "bob"},
		{Directive: protocol.DeleteDirective, Key: "user:1", Value: ""},
		{Directive: protocol.SetDirective, Key: "user:3", Value: "charlie"},
	}

	for _, cmd := range commands {
		if err := wal1.Append(cmd); err != nil {
			t.Fatalf("Append() error = %v", err)
		}
	}
	wal1.Close()

	// Now replay in a new WAL instance
	wal2, err := NewWAL(testPath)
	if err != nil {
		t.Fatalf("NewWAL() error = %v", err)
	}
	defer wal2.Close()

	s := store.NewStore()
	err = wal2.Replay(s)
	if err != nil {
		t.Errorf("Replay() error = %v", err)
	}

	// Verify final state
	if _, exists := s.Get("user:1"); exists {
		t.Error("user:1 should not exist after DELETE")
	}
	if value, exists := s.Get("user:2"); !exists || value != "bob" {
		t.Errorf("user:2 = %q, want 'bob'", value)
	}
	if value, exists := s.Get("user:3"); !exists || value != "charlie" {
		t.Errorf("user:3 = %q, want 'charlie'", value)
	}
}

func TestReplayWithInvalidCommand(t *testing.T) {
	tempDir := t.TempDir()
	testPath := filepath.Join(tempDir, "test.log")

	// Write an invalid command
	file, err := os.Create(testPath)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file.WriteString("INVALID foo bar\n")
	file.Close()

	wal, err := NewWAL(testPath)
	if err != nil {
		t.Fatalf("NewWAL() error = %v", err)
	}
	defer wal.Close()

	s := store.NewStore()
	err = wal.Replay(s)
	if err == nil {
		t.Error("Replay() should return error for invalid command")
	}
}

func TestReplayNonExistentFile(t *testing.T) {
	tempDir := t.TempDir()
	testPath := filepath.Join(tempDir, "nonexistent.log")

	// Don't create the file
	wal, err := NewWAL(testPath)
	if err != nil {
		t.Fatalf("NewWAL() error = %v", err)
	}
	defer wal.Close()

	s := store.NewStore()
	err = wal.Replay(s)
	// Should not error - just means no data to replay yet
	if err != nil {
		t.Errorf("Replay() on non-existent file should not error, got: %v", err)
	}
}
