package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Entry is the base struct that all stored items should embed.
type Entry struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

// NewEntry creates an Entry with a timestamp-based ID and current time.
func NewEntry() Entry {
	return Entry{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		CreatedAt: time.Now().UTC(),
	}
}

// Store is a generic, JSON-file-backed store for any type T.
type Store[T any] struct {
	path string
}

// NewStore creates a Store backed by a JSON file at the given path.
// The directory is created on first use (see ensureDir).
func NewStore[T any](path string) *Store[T] {
	return &Store[T]{path: path}
}

func (s *Store[T]) ensureDir() error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("store: cannot create directory %s: %w", dir, err)
	}
	return nil
}

// Load reads all entries from the JSON file.
// Returns an empty slice (not an error) if the file does not exist yet.
func (s *Store[T]) Load() ([]T, error) {
	if err := s.ensureDir(); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return []T{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("store: read %s: %w", s.path, err)
	}

	var items []T
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("store: decode %s: %w", s.path, err)
	}
	return items, nil
}

// Save overwrites the JSON file with the given slice.
func (s *Store[T]) Save(items []T) error {
	if err := s.ensureDir(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return fmt.Errorf("store: encode: %w", err)
	}

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("store: write temp file: %w", err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		return fmt.Errorf("store: rename temp file: %w", err)
	}
	return nil
}

// Append loads existing entries, appends item, and saves.
func (s *Store[T]) Append(item T) error {
	items, err := s.Load()
	if err != nil {
		return err
	}
	items = append(items, item)
	return s.Save(items)
}
