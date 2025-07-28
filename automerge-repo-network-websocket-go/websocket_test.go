package network_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/automerge/automerge-repo-go"
	"github.com/automerge/automerge-repo-network-websocket-go"
	"github.com/google/uuid"
)

func TestWebSocketHandshake(t *testing.T) {
	serverRepo := repo.New()
	clientRepo := repo.New()

	var remoteFromServer repo.RepoID
	var wsErr error

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, remote, err := network.AcceptWebSocket(w, r, serverRepo.ID)
		if err != nil {
			wsErr = err
			return
		}
		remoteFromServer = remote
		conn.Close()
	}))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	conn, remoteFromClient, err := network.DialWebSocket(ctx, wsURL, clientRepo.ID)
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}
	conn.Close()

	if wsErr != nil {
		t.Fatalf("accept error: %v", wsErr)
	}

	if remoteFromClient != serverRepo.ID || remoteFromServer != clientRepo.ID {
		t.Fatalf("unexpected repo IDs: %v %v", remoteFromClient, remoteFromServer)
	}
}

func TestWSConnSendRecvMessage(t *testing.T) {
	serverRepo := repo.New()
	clientRepo := repo.New()

	var received repo.RepoMessage
	done := make(chan struct{})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, _, err := network.AcceptWebSocket(w, r, serverRepo.ID)
		if err != nil {
			t.Errorf("accept error: %v", err)
			close(done)
			return
		}
		defer conn.Close()
		received, err = conn.RecvMessage()
		if err != nil {
			t.Errorf("recv error: %v", err)
		}
		close(done)
	}))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	conn, _, err := network.DialWebSocket(ctx, wsURL, clientRepo.ID)
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}
	defer conn.Close()

	msg := repo.RepoMessage{
		Type:       "sync",
		FromRepoID: clientRepo.ID,
		ToRepoID:   serverRepo.ID,
		DocumentID: uuid.New(),
		Message:    []byte("ws"),
	}
	if err := conn.SendMessage(msg); err != nil {
		t.Fatalf("send error: %v", err)
	}

	<-done

	if received.Type != msg.Type || received.FromRepoID != msg.FromRepoID || received.ToRepoID != msg.ToRepoID || received.DocumentID != msg.DocumentID || string(received.Message) != string(msg.Message) {
		t.Fatalf("mismatch: %#v vs %#v", received, msg)
	}
}
