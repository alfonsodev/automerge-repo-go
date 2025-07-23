package repo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

func TestFsStoreSaveLoadList(t *testing.T) {
	dir := t.TempDir()
	store := &FsStore{Dir: dir}

	// create document
	doc := &Document{ID: uuid.New(), Data: map[string]interface{}{"foo": "bar"}}
	if err := store.Save(doc); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// file should exist
	path := filepath.Join(dir, doc.ID.String()+".json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file %s to exist: %v", path, err)
	}

	// load document
	loaded, err := store.Load(doc.ID)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if v, ok := loaded.Data["foo"]; !ok || v != "bar" {
		t.Fatalf("unexpected loaded data: %v", loaded.Data)
	}

	// list ids
	ids, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(ids) != 1 || ids[0] != doc.ID {
		t.Fatalf("unexpected ids: %v", ids)
	}
}
