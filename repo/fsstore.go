package repo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
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
	var data map[string]interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, err
	}
	return &Document{ID: id, Data: data}, nil
}

// List returns all document IDs currently stored on disk.
func (s *FsStore) List() ([]DocumentID, error) {
	var ids []DocumentID
	files, err := os.ReadDir(s.Dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return ids, nil
		}
		return nil, err
	}
	for _, f := range files {
		name := f.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		idStr := strings.TrimSuffix(name, ".json")
		id, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids, nil
}
