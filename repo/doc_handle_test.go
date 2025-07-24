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
