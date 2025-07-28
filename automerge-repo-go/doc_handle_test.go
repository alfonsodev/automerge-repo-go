package repo

import (
	"fmt"
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

func ExampleRepo_NewDocHandle() {
	// Create a new repo.
	r := New()

	// Create a new document handle.
	h := r.NewDocHandle()

	// Mutate the document.
	err := h.WithDocMut(func(doc *automerge.Doc) error {
		return doc.RootMap().Set("foo", "bar")
	})
	if err != nil {
		panic(err)
	}

	// Read the value back.
	var val string
	h.WithDoc(func(doc *automerge.Doc) {
		v, err := doc.RootMap().Get("foo")
		if err != nil {
			panic(err)
		}
		val = v.Str()
	})
	fmt.Println(val)
	// Output: bar
}
