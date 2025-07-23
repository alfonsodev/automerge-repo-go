package repo

import (
	"testing"
)

func TestRepoSaveLoadDoc(t *testing.T) {
	storeDir := t.TempDir()
	store := &FsStore{Dir: storeDir}
	r := NewWithStore(store)

	doc := r.NewDoc()
	if err := doc.Set("name", "Alice"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	if err := r.SaveDoc(doc.ID); err != nil {
		t.Fatalf("SaveDoc failed: %v", err)
	}

	// clear repo memory
	r.docs = make(map[DocumentID]*Document)

	loaded, err := r.LoadDoc(doc.ID)
	if err != nil {
		t.Fatalf("LoadDoc failed: %v", err)
	}
	if v, _ := loaded.Get("name"); v != "Alice" {
		t.Fatalf("unexpected value: %v", v)
	}
}
