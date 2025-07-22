package repo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// FsStore persists documents to disk in a directory.
type FsStore struct {
	Dir string
}

// Save writes the given document to the filesystem.
func (s *FsStore) Save(doc *Document) error {
	if err := os.MkdirAll(s.Dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(s.Dir, fmt.Sprintf("%s.json", doc.ID))
	data, err := json.Marshal(doc.Data)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// Load reads a document from disk.
func (s *FsStore) Load(id DocumentID) (*Document, error) {
	path := filepath.Join(s.Dir, fmt.Sprintf("%s.json", id))
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("document %s not found", id)
		}
		return nil, err
	}
	var data []byte
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, err
	}
	return &Document{ID: id, Data: data}, nil
}
