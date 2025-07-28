package repo

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestConnect(t *testing.T) {
	c1, c2 := net.Pipe()
	repo1 := New()
	repo2 := New()

	var remote1 RepoID
	var remote2 RepoID
	var lp1 *LPConn
	var lp2 *LPConn
	errCh := make(chan error, 2)
	ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second)
	defer cancel1()
	go func() {
		var err error
		lp1, remote1, err = Connect(ctx1, c1, repo1.ID, Outgoing)
		errCh <- err
	}()
	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second)
	defer cancel2()
	go func() {
		var err error
		lp2, remote2, err = Connect(ctx2, c2, repo2.ID, Incoming)
		errCh <- err
	}()
	if err := <-errCh; err != nil {
		t.Fatalf("connect error: %v", err)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("connect error: %v", err)
	}
	if remote1 != repo2.ID || remote2 != repo1.ID {
		t.Fatalf("unexpected repo IDs: %v %v", remote1, remote2)
	}
	lp1.Close()
	lp2.Close()
}

func TestLPConnSendRecvMessage(t *testing.T) {
	c1, c2 := net.Pipe()
	lp1 := NewLPConn(c1)
	lp2 := NewLPConn(c2)

	msg := RepoMessage{
		Type:       "sync",
		FromRepoID: New().ID,
		ToRepoID:   New().ID,
		DocumentID: uuid.New(),
		Message:    []byte("hi"),
	}

	go func() {
		if err := lp1.SendMessage(msg); err != nil {
			t.Errorf("send error: %v", err)
		}
	}()

	got, err := lp2.RecvMessage()
	if err != nil {
		t.Fatalf("recv error: %v", err)
	}
	if got.Type != msg.Type || got.FromRepoID != msg.FromRepoID || got.ToRepoID != msg.ToRepoID || got.DocumentID != msg.DocumentID || string(got.Message) != string(msg.Message) {
		t.Fatalf("mismatch: %#v vs %#v", got, msg)
	}
	lp1.Close()
	lp2.Close()
}
