package repo

import "testing"

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
	if err := cc.Await(); err == nil {
		t.Fatalf("expected error, got nil")
	}

	// drain events
	<-h1.Events
	<-h1.Events

	h1.Close()
	h2.Close()
}
