package store

import (
	"testing"
)

func TestNewStore(t *testing.T) {
	s := NewStore()
	if s == nil {
		t.Fatal("NewStore() returned nil")
	}
	if s.data == nil {
		t.Error("NewStore() did not initialize data map")
	}
}

func TestSetAndGet(t *testing.T) {
	s := NewStore()

	// Set a value
	s.Set("mykey", "myvalue")

	// Get it back
	value, exists := s.Get("mykey")
	if !exists {
		t.Error("Get() returned exists=false for key that was just set")
	}
	if value != "myvalue" {
		t.Errorf("Get() = %q, want %q", value, "myvalue")
	}
}

func TestGetNonExistent(t *testing.T) {
	s := NewStore()

	value, exists := s.Get("nonexistent")
	if exists {
		t.Error("Get() returned exists=true for non-existent key")
	}
	if value != "" {
		t.Errorf("Get() returned value %q for non-existent key, want empty string", value)
	}
}

func TestSetOverwrite(t *testing.T) {
	s := NewStore()

	// Set initial value
	s.Set("mykey", "value1")

	// Overwrite it
	s.Set("mykey", "value2")

	// Should get the new value
	value, exists := s.Get("mykey")
	if !exists {
		t.Error("Get() returned exists=false after overwrite")
	}
	if value != "value2" {
		t.Errorf("Get() = %q after overwrite, want %q", value, "value2")
	}
}

func TestDelete(t *testing.T) {
	s := NewStore()

	// Set a value
	s.Set("mykey", "myvalue")

	// Delete it
	deleted := s.Delete("mykey")
	if !deleted {
		t.Error("Delete() returned false for existing key")
	}

	// Verify it's gone
	_, exists := s.Get("mykey")
	if exists {
		t.Error("Get() returned exists=true after Delete()")
	}
}

func TestDeleteNonExistent(t *testing.T) {
	s := NewStore()

	deleted := s.Delete("nonexistent")
	if deleted {
		t.Error("Delete() returned true for non-existent key")
	}
}

func TestDeleteIdempotent(t *testing.T) {
	s := NewStore()

	// Set and delete
	s.Set("mykey", "myvalue")
	s.Delete("mykey")

	// Delete again
	deleted := s.Delete("mykey")
	if deleted {
		t.Error("Delete() returned true on second delete of same key")
	}
}

func TestExists(t *testing.T) {
	s := NewStore()

	// Key doesn't exist yet
	if s.Exists("mykey") {
		t.Error("Exists() returned true for non-existent key")
	}

	// Set the key
	s.Set("mykey", "myvalue")

	// Now it should exist
	if !s.Exists("mykey") {
		t.Error("Exists() returned false for existing key")
	}

	// Delete it
	s.Delete("mykey")

	// Should not exist anymore
	if s.Exists("mykey") {
		t.Error("Exists() returned true after Delete()")
	}
}

func TestEmptyValue(t *testing.T) {
	s := NewStore()

	// Set empty string as value
	s.Set("mykey", "")

	// Should exist with empty value
	value, exists := s.Get("mykey")
	if !exists {
		t.Error("Get() returned exists=false for key with empty value")
	}
	if value != "" {
		t.Errorf("Get() = %q, want empty string", value)
	}

	// Should report as existing
	if !s.Exists("mykey") {
		t.Error("Exists() returned false for key with empty value")
	}
}

func TestEmptyKey(t *testing.T) {
	s := NewStore()

	// Empty key is technically valid in a map
	s.Set("", "value")

	value, exists := s.Get("")
	if !exists {
		t.Error("Get() returned exists=false for empty key")
	}
	if value != "value" {
		t.Errorf("Get() = %q for empty key, want %q", value, "value")
	}
}

func TestMultipleKeys(t *testing.T) {
	s := NewStore()

	// Set multiple keys
	keys := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	for k, v := range keys {
		s.Set(k, v)
	}

	// Verify all exist with correct values
	for k, want := range keys {
		got, exists := s.Get(k)
		if !exists {
			t.Errorf("Get(%q) returned exists=false", k)
		}
		if got != want {
			t.Errorf("Get(%q) = %q, want %q", k, got, want)
		}
	}

	// Delete one
	s.Delete("key2")

	// Verify key1 and key3 still exist, key2 doesn't
	if _, exists := s.Get("key1"); !exists {
		t.Error("key1 should still exist after deleting key2")
	}
	if _, exists := s.Get("key2"); exists {
		t.Error("key2 should not exist after delete")
	}
	if _, exists := s.Get("key3"); !exists {
		t.Error("key3 should still exist after deleting key2")
	}
}

func TestSpecialCharacterKeys(t *testing.T) {
	s := NewStore()

	specialKeys := []string{
		"user:123:name",
		"cache/item/abc",
		"key.with.dots",
		"key-with-dashes",
		"key_with_underscores",
		"key with spaces",
		"key\twith\ttabs",
	}

	for _, key := range specialKeys {
		s.Set(key, "value")
		value, exists := s.Get(key)
		if !exists {
			t.Errorf("Get(%q) returned exists=false", key)
		}
		if value != "value" {
			t.Errorf("Get(%q) = %q, want %q", key, value, "value")
		}
	}
}
