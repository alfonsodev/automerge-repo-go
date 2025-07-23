package repo

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	automerge "github.com/automerge/automerge-go"
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
	path := filepath.Join(s.Dir, fmt.Sprintf("%s.automerge", doc.ID))
	if doc.doc == nil {
		doc.doc = automerge.New()
	}
	data := doc.doc.Save()
	return os.WriteFile(path, data, 0o644)
}

// Load reads a document from disk.
func (s *FsStore) Load(id DocumentID) (*Document, error) {
	path := filepath.Join(s.Dir, fmt.Sprintf("%s.automerge", id))
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("document %s not found", id)
		}
		return nil, err
	}
	d, err := automerge.Load(b)
	if err != nil {
		return nil, err
	}
	return &Document{ID: id, doc: d}, nil
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
		if !strings.HasSuffix(name, ".automerge") {
			continue
		}
		idStr := strings.TrimSuffix(name, ".automerge")
		id, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids, nil
}
