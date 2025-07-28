package repo

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	automerge "github.com/automerge/automerge-go"
)

func TestDirectSync(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	h1 := NewRepoHandle(New())
	h2 := NewRepoHandle(New())

	// Connect the two repo handles
	c1, c2 := net.Pipe()
	var wg sync.WaitGroup
	var ca, cb *LPConn
	var ra, rb RepoID
	var errA, errB error
	wg.Add(2)
	go func() {
		defer wg.Done()
		ca, rb, errA = Connect(ctx, c1, h1.Repo.ID, Outgoing)
	}()
	go func() {
		defer wg.Done()
		cb, ra, errB = Connect(ctx, c2, h2.Repo.ID, Incoming)
	}()
	wg.Wait()
	if errA != nil || errB != nil {
		t.Fatalf("connect error: %v %v", errA, errB)
	}
	_ = h1.AddConn(rb, ca)
	_ = h2.AddConn(ra, cb)

	// Create a document in h1 and make a change
	doc1 := h1.Repo.NewDoc()
	_ = doc1.Set("key", "value")

	// Ensure h2 has the document (it should be created on first sync message)
	doc2, _ := h2.Repo.GetDoc(doc1.ID)
	if doc2 == nil {
		doc2 = &Document{ID: doc1.ID, Doc: automerge.New()}
		h2.Repo.docs[doc1.ID] = doc2
	}

	// Sync from h1 to h2
	if err := h1.SyncAll(h2.Repo.ID); err != nil {
		t.Fatalf("SyncAll error: %v", err)
	}

	// Give some time for sync messages to be processed
	time.Sleep(500 * time.Millisecond)

	// Verify the document in h2
	if v, ok := doc2.Get("key"); !ok || v != "value" {
		t.Fatalf("document in h2 not synced: %v %v", v, ok)
	}

	h1.Close()
	h2.Close()
}