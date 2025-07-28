package repo

import (
	"testing"
	"time"

	automerge "github.com/automerge/automerge-go"
)

// newMockConn reused from handle_test.go

func TestRepoHandleSync(t *testing.T) {
	h1 := NewRepoHandle(New())
	h2 := NewRepoHandle(New())

	c1, c2 := newMockConn()
	_ = h1.AddConn(h2.Repo.ID, c1)
	_ = h2.AddConn(h1.Repo.ID, c2)

	doc1 := h1.Repo.NewDoc()
	if err := doc1.Set("k", "v"); err != nil {
		t.Fatalf("set err: %v", err)
	}
	doc2 := &Document{ID: doc1.ID, doc: automerge.New()}
	h2.Repo.docs[doc1.ID] = doc2

	if err := h1.SyncDocument(h2.Repo.ID, doc1.ID); err != nil {
		t.Fatalf("sync error: %v", err)
	}

	// give the goroutines a moment to process
	time.Sleep(10 * time.Millisecond)

	if v, ok := doc2.Get("k"); !ok || v != "v" {
		t.Fatalf("doc not synced: %v %v", v, ok)
	}

	h1.Close()
	h2.Close()
}
