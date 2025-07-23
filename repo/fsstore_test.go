package repo

import (
	"os"
	"path/filepath"
	"testing"

	automerge "github.com/automerge/automerge-go"
	"github.com/google/uuid"
)

func TestFsStoreSaveLoadList(t *testing.T) {
	dir := t.TempDir()
	store := &FsStore{Dir: dir}

	// create document
	doc := &Document{ID: uuid.New(), doc: automerge.New()}
	if err := doc.Set("foo", "bar"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	if err := store.Save(doc); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// file should exist
	path := filepath.Join(dir, doc.ID.String()+".automerge")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file %s to exist: %v", path, err)
	}

	// load document
	loaded, err := store.Load(doc.ID)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if v, ok := loaded.Get("foo"); !ok || v != "bar" {
		t.Fatalf("unexpected loaded data: %v", v)
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
