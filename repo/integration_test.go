package repo

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"
)

func connectHandles(t *testing.T, ctx context.Context, a, b *RepoHandle) (ConnComplete, ConnComplete) {
	c1, c2 := net.Pipe()
	var wg sync.WaitGroup
	var ca *LPConn
	var cb *LPConn
	var ra RepoID
	var rb RepoID
	var errA, errB error
	wg.Add(2)
	go func() {
		defer wg.Done()
		ca, rb, errA = Connect(ctx, c1, a.Repo.ID, Outgoing)
	}()
	go func() {
		defer wg.Done()
		cb, ra, errB = Connect(ctx, c2, b.Repo.ID, Incoming)
	}()
	wg.Wait()
	if errA != nil || errB != nil {
		t.Fatalf("connect error: %v %v", errA, errB)
	}
	cca := a.AddConn(rb, ca)
	ccb := b.AddConn(ra, cb)
	if evt := <-a.Events; evt.Type != EventPeerConnected || evt.Peer != rb {
		t.Fatalf("expected peer connected event, got %#v", evt)
	}
	if evt := <-b.Events; evt.Type != EventPeerConnected || evt.Peer != ra {
		t.Fatalf("expected peer connected event, got %#v", evt)
	}
	return cca, ccb
}

func TestMultiPeerSync(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	h1 := NewRepoHandle(New())
	h2 := NewRepoHandle(New())
	h3 := NewRepoHandle(New())

	cc12a, cc12b := connectHandles(t, ctx, h1, h2)
	cc23a, cc23b := connectHandles(t, ctx, h2, h3)
	cc31a, cc31b := connectHandles(t, ctx, h3, h1)

	doc := h1.Repo.NewDoc()
	if err := doc.Set("k", "v"); err != nil {
		t.Fatalf("set err: %v", err)
	}
	if err := h1.SyncAll(h2.Repo.ID); err != nil {
		t.Fatalf("sync err: %v", err)
	}
	if err := h1.SyncAll(h3.Repo.ID); err != nil {
		t.Fatalf("sync err: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	if d, ok := h2.Repo.GetDoc(doc.ID); !ok {
		t.Fatalf("h2 missing doc")
	} else if v, ok := d.Get("k"); !ok || v != "v" {
		t.Fatalf("h2 doc mismatch: %v %v", v, ok)
	}
	if d, ok := h3.Repo.GetDoc(doc.ID); !ok {
		t.Fatalf("h3 missing doc")
	} else if v, ok := d.Get("k"); !ok || v != "v" {
		t.Fatalf("h3 doc mismatch: %v %v", v, ok)
	}

	h1.Close()
	h2.Close()
	h3.Close()
	_ = cc12a.Await()
	_ = cc12b.Await()
	_ = cc23a.Await()
	_ = cc23b.Await()
	_ = cc31a.Await()
	_ = cc31b.Await()
}
