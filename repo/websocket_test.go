package repo

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWebSocketHandshake(t *testing.T) {
	serverRepo := New()
	clientRepo := New()

	var remoteFromServer RepoID
	var wsErr error

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, remote, err := AcceptWebSocket(w, r, serverRepo.ID)
		if err != nil {
			wsErr = err
			return
		}
		remoteFromServer = remote
		conn.Close()
	}))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")

	conn, remoteFromClient, err := DialWebSocket(wsURL, clientRepo.ID)
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
