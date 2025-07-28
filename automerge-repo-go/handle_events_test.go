package repo

import (
	"context"
	"testing"
	"time"
)

func TestRepoHandleConnectionEvents(t *testing.T) {
	h1 := NewRepoHandle(New())
	h2 := NewRepoHandle(New())

	c1, _ := newMockConn()

	go func() {
		_ = h1.AddConn(h2.Repo.ID, c1)
	}()

	if evt := <-h1.Events; evt.Type != EventPeerConnected || evt.Peer != h2.Repo.ID {
		t.Fatalf("expected peer connected event, got %#v", evt)
	}

	h1.RemoveConn(h2.Repo.ID)
	if evt := <-h1.Events; evt.Type != EventPeerDisconnected || evt.Peer != h2.Repo.ID {
		t.Fatalf("expected peer disconnected event, got %#v", evt)
	}

	h1.Close()
	h2.Close()
}

func TestRepoHandleConnErrorEvent(t *testing.T) {
	h1 := NewRepoHandle(New())
	h2 := NewRepoHandle(New())

	c1, c2 := newMockConn()

	cc := h1.AddConn(h2.Repo.ID, c1)
	if evt := <-h1.Events; evt.Type != EventPeerConnected || evt.Peer != h2.Repo.ID {
		t.Fatalf("expected peer connected event, got %#v", evt)
	}

	// simulate remote closing connection
	c2.Close()

	evt := <-h1.Events
	if evt.Type != EventConnError || evt.Peer != h2.Repo.ID || evt.Err == nil {
		t.Fatalf("expected conn error event, got %#v", evt)
	}

	evt = <-h1.Events
	if evt.Type != EventPeerDisconnected || evt.Peer != h2.Repo.ID {
		t.Fatalf("expected peer disconnected event, got %#v", evt)
	}

	h1.Close()
	h2.Close()

	// consume completion to avoid goroutine leak
	_ = cc.Await()
}

func TestRepoHandleConnComplete(t *testing.T) {
	h1 := NewRepoHandle(New())
	h2 := NewRepoHandle(New())

	c1, c2 := newMockConn()
	cc := h1.AddConn(h2.Repo.ID, c1)
	_ = h2.AddConn(h1.Repo.ID, c2)

	if evt := <-h1.Events; evt.Type != EventPeerConnected || evt.Peer != h2.Repo.ID {
		t.Fatalf("expected peer connected event, got %#v", evt)
	}

	// close the connection from remote
	c2.Close()

	// wait for completion
	if fin := cc.Await(); fin.Err == nil {
		t.Fatalf("expected error, got nil")
	}

	// drain events
	<-h1.Events
	<-h1.Events

	h1.Close()
	h2.Close()
}

func TestRepoHandleReconnect(t *testing.T) {
	h1 := NewRepoHandle(New())
	h2 := NewRepoHandle(New())

	rc := make(chan *mockConn, 2)
	dial := func(ctx context.Context) (Conn, error) {
		c1, c2 := newMockConn()
		_ = h2.AddConn(h1.Repo.ID, c2)
		rc <- c2
		return c1, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	cc := h1.AddConnWithRetry(ctx, h2.Repo.ID, dial, 10*time.Millisecond)

	if evt := <-h1.Events; evt.Type != EventPeerConnected || evt.Peer != h2.Repo.ID {
		t.Fatalf("expected peer connected event, got %#v", evt)
	}
	first := <-rc
	first.Close()

	<-h1.Events // conn_error
	if evt := <-h1.Events; evt.Type != EventPeerDisconnected {
		t.Fatalf("expected peer disconnected, got %#v", evt)
	}

	if evt := <-h1.Events; evt.Type != EventPeerConnected {
		t.Fatalf("expected peer reconnected, got %#v", evt)
	}
	second := <-rc
	cancel()
	second.Close()
	_ = cc.Await()

	h1.Close()
	h2.Close()
}
