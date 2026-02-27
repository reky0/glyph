package cmd

import (
	"fmt"
	"path/filepath"

	core "github.com/reky0/glyph-core"
	store "github.com/reky0/glyph-store"
)

func openStore() (*store.Store[PinEntry], error) {
	paths := core.NewPaths("pin")
	dir, err := paths.DataDir()
	if err != nil {
		return nil, err
	}
	return store.NewStore[PinEntry](filepath.Join(dir, "pins.json")), nil
}

func loadEntries() ([]PinEntry, *store.Store[PinEntry], error) {
	s, err := openStore()
	if err != nil {
		return nil, nil, err
	}
	entries, err := s.Load()
	if err != nil {
		return nil, nil, err
	}
	return entries, s, nil
}

func findByID(entries []PinEntry, id string) (PinEntry, int, error) {
	for i, e := range entries {
		if e.ID == id || shortID(e.ID) == id {
			return e, i, nil
		}
	}
	return PinEntry{}, -1, fmt.Errorf("no entry with id %q", id)
}
