package repo

import (
	"testing"

	automerge "github.com/automerge/automerge-go"
)

func TestDocumentHandleChanged(t *testing.T) {
	r := New()
	h := r.NewDocHandle()

	ch := h.Changed()

	err := h.WithDocMut(func(doc *automerge.Doc) error {
		doc.RootMap().Set("k", "v")
		return nil
	})
	if err != nil {
		t.Fatalf("mutate err: %v", err)
	}

	select {
	case <-ch:
	default:
		t.Fatalf("expected change notification")
	}
}

func TestDocumentHandleAutoSave(t *testing.T) {
	dir := t.TempDir()
	r := NewWithStore(&FsStore{Dir: dir})
	h := r.NewDocHandle()

	if err := h.WithDocMut(func(doc *automerge.Doc) error {
		doc.RootMap().Set("foo", "bar")
		return nil
	}); err != nil {
		t.Fatalf("mutate err: %v", err)
	}

	r2 := NewWithStore(&FsStore{Dir: dir})
	loaded, err := r2.LoadDoc(h.doc.ID)
	if err != nil {
		t.Fatalf("load err: %v", err)
	}
	if v, _ := loaded.Get("foo"); v != "bar" {
		t.Fatalf("unexpected value: %v", v)
	}
}
