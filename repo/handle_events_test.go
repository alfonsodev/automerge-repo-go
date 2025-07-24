package repo

import "testing"

func TestRepoHandleConnectionEvents(t *testing.T) {
	h1 := NewRepoHandle(New())
	h2 := NewRepoHandle(New())

	c1, _ := newMockConn()

	go func() {
		h1.AddConn(h2.Repo.ID, c1)
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

	h1.AddConn(h2.Repo.ID, c1)
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
}

func TestRepoHandleConnComplete(t *testing.T) {
	h1 := NewRepoHandle(New())
	h2 := NewRepoHandle(New())

	c1, c2 := newMockConn()

	cc := h1.AddConn(h2.Repo.ID, c1)
	if evt := <-h1.Events; evt.Type != EventPeerConnected || evt.Peer != h2.Repo.ID {
		t.Fatalf("expected peer connected event, got %#v", evt)
	}

	c2.Close()

	<-h1.Events // conn_error
	<-h1.Events // peer_disconnected

	if reason := cc.Wait(); reason != ConnFinishedTheyDisconnected {
		t.Fatalf("unexpected reason: %v", reason)
	}

	h1.Close()
	h2.Close()
}
