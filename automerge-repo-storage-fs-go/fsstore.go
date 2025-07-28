package storage

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	automerge "github.com/automerge/automerge-go"
	"github.com/automerge/automerge-repo-go"
	"github.com/google/uuid"
)

// FsStore persists documents to disk in a directory.
type FsStore struct {
	Dir string
}

// Save appends any new changes from the document to a file on disk.
// If the file does not exist, it creates a new one with a full snapshot of the document.
func (s *FsStore) Save(doc *repo.Document) error {
	if err := os.MkdirAll(s.Dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(s.Dir, fmt.Sprintf("%s.automerge", doc.ID))

	if doc.Doc == nil {
		doc.Doc = automerge.New()
	}

	// If the file doesn't exist, save the full document.
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return s.Compact(doc)
	}

	data := doc.Doc.SaveIncremental()

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return err
	}

	return nil
}

// Compact writes the full document to disk, replacing any incremental saves.
func (s *FsStore) Compact(doc *repo.Document) error {
	if err := os.MkdirAll(s.Dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(s.Dir, fmt.Sprintf("%s.automerge", doc.ID))
	if doc.Doc == nil {
		doc.Doc = automerge.New()
	}
	data := doc.Doc.Save()
	return os.WriteFile(path, data, 0o644)
}

// Load reads a document from disk. It can load both full snapshots and files
// with incremental changes appended.
func (s *FsStore) Load(id repo.DocumentID) (*repo.Document, error) {
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
	return &repo.Document{ID: id, Doc: d}, nil
}

// List returns all document IDs currently stored on disk.
func (s *FsStore) List() ([]repo.DocumentID, error) {
	var ids []repo.DocumentID
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