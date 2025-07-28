package storage_test

import (
	"testing"

	"github.com/automerge/automerge-repo-go"
	"github.com/automerge/automerge-repo-storage-fs-go"
	automerge "github.com/automerge/automerge-go"
)

func TestDocumentHandleAutoSave(t *testing.T) {
	dir := t.TempDir()
	r := repo.NewWithStore(&storage.FsStore{Dir: dir})
	h := r.NewDocHandle()

	if err := h.WithDocMut(func(doc *automerge.Doc) error {
		doc.RootMap().Set("foo", "bar")
		return nil
	}); err != nil {
		t.Fatalf("mutate err: %v", err)
	}

	if err := h.Save(); err != nil {
		t.Fatalf("save err: %v", err)
	}

	r2 := repo.NewWithStore(&storage.FsStore{Dir: dir})
	loaded, err := r2.LoadDoc(h.DocID())
	if err != nil {
		t.Fatalf("load err: %v", err)
	}
	if v, _ := loaded.Get("foo"); v != "bar" {
		t.Fatalf("unexpected value: %v", v)
	}
}

func TestRepoSaveLoadDoc(t *testing.T) {
	storeDir := t.TempDir()
	store := &storage.FsStore{Dir: storeDir}
	r := repo.NewWithStore(store)

	doc := r.NewDoc()
	if err := doc.Set("name", "Alice"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	if err := r.SaveDoc(doc.ID); err != nil {
		t.Fatalf("SaveDoc failed: %v", err)
	}

	// clear repo memory
	r.ClearDocs()

	loaded, err := r.LoadDoc(doc.ID)
	if err != nil {
		t.Fatalf("LoadDoc failed: %v", err)
	}
	if v, _ := loaded.Get("name"); v != "Alice" {
		t.Fatalf("unexpected value: %v", v)
	}
}