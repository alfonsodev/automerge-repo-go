package repo

import (
	"fmt"
	"io"
	"testing"
	"time"
)

// mockConn is a simple in-memory connection for tests.
type mockConn struct {
	sendCh chan RepoMessage
	recvCh chan RepoMessage
}

// sendErrConn fails on SendMessage but otherwise behaves like an idle connection.
type sendErrConn struct {
	recvCh chan RepoMessage
}

func newSendErrConn() *sendErrConn {
	return &sendErrConn{recvCh: make(chan RepoMessage)}
}

func (c *sendErrConn) SendMessage(m RepoMessage) error { return fmt.Errorf("send fail") }

func (c *sendErrConn) RecvMessage() (RepoMessage, error) {
	msg, ok := <-c.recvCh
	if !ok {
		return RepoMessage{}, io.EOF
	}
	return msg, nil
}

func (c *sendErrConn) Close() error { close(c.recvCh); return nil }

func newMockConn() (*mockConn, *mockConn) {
	c1 := &mockConn{sendCh: make(chan RepoMessage, 1), recvCh: make(chan RepoMessage, 1)}
	c2 := &mockConn{sendCh: c1.recvCh, recvCh: c1.sendCh}
	return c1, c2
}

func (c *mockConn) SendMessage(m RepoMessage) error {
	c.sendCh <- m
	return nil
}

func (c *mockConn) RecvMessage() (RepoMessage, error) {
	msg, ok := <-c.recvCh
	if !ok {
		return RepoMessage{}, io.EOF
	}
	return msg, nil
}

func (c *mockConn) Close() error {
	close(c.sendCh)
	return nil
}

func TestRepoHandleMessageForwarding(t *testing.T) {
	h1 := NewRepoHandle(New())
	h2 := NewRepoHandle(New())

	c1, c2 := newMockConn()
	h1.AddConn(h2.Repo.ID, c1)
	h2.AddConn(h1.Repo.ID, c2)

	msg := RepoMessage{Type: "ephemeral", FromRepoID: h1.Repo.ID, ToRepoID: h2.Repo.ID}
	if err := h1.SendMessage(h2.Repo.ID, msg); err != nil {
		t.Fatalf("SendMessage error: %v", err)
	}

	select {
	case got := <-h2.Inbox:
		if got.Type != msg.Type || got.FromRepoID != msg.FromRepoID || got.ToRepoID != msg.ToRepoID {
			t.Fatalf("unexpected message: %#v", got)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for message")
	}

	h1.Close()
	h2.Close()
}

func TestRepoHandleSendErrorEvent(t *testing.T) {
	h := NewRepoHandle(New())
	remoteID := New().ID

	c := newSendErrConn()
	h.AddConn(remoteID, c)

	if evt := <-h.Events; evt.Type != EventPeerConnected || evt.Peer != remoteID {
		t.Fatalf("expected peer connected event, got %#v", evt)
	}

	msg := RepoMessage{Type: "ephemeral", FromRepoID: h.Repo.ID, ToRepoID: remoteID}
	if err := h.SendMessage(remoteID, msg); err == nil {
		t.Fatal("expected send error")
	}

	evt := <-h.Events
	if evt.Type != EventConnError || evt.Peer != remoteID {
		t.Fatalf("expected conn error event, got %#v", evt)
	}

	evt = <-h.Events
	if evt.Type != EventPeerDisconnected || evt.Peer != remoteID {
		t.Fatalf("expected peer disconnected event, got %#v", evt)
	}

	h.Close()
}
