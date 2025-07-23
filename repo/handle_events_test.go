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
