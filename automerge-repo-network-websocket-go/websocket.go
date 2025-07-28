package network

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/automerge/automerge-repo-go"
	"github.com/fxamacker/cbor/v2"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type handshakeMessage struct {
	Type     string `cbor:"type"`
	SenderID string `cbor:"senderId"`
}

// WSConn wraps a websocket connection for sending CBOR messages.
type WSConn struct {
	c  *websocket.Conn
	mu sync.Mutex
}

// NewWSConn creates a new WSConn.
func NewWSConn(c *websocket.Conn) *WSConn {
	return &WSConn{c: c}
}

// Send encodes v as CBOR and writes it over the websocket.
func (c *WSConn) Send(v interface{}) error {
	data, err := cbor.Marshal(v)
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.WriteMessage(websocket.BinaryMessage, data)
}

// Recv reads a CBOR message into v.
func (c *WSConn) Recv(v interface{}) error {
	_, data, err := c.c.ReadMessage()
	if err != nil {
		return err
	}
	return cbor.Unmarshal(data, v)
}

// SendMessage sends a repo.RepoMessage as CBOR over the websocket.
func (c *WSConn) SendMessage(msg repo.RepoMessage) error {
	data, err := msg.Encode()
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.WriteMessage(websocket.BinaryMessage, data)
}

// RecvMessage reads a repo.RepoMessage from the websocket.
func (c *WSConn) RecvMessage() (repo.RepoMessage, error) {
	_, data, err := c.c.ReadMessage()
	if err != nil {
		return repo.RepoMessage{}, err
	}
	return repo.DecodeRepoMessage(data)
}

// Close closes the websocket.
func (c *WSConn) Close() error { return c.c.Close() }

// DialWebSocket dials the given websocket URL and performs the join/peer handshake.
// It returns the remote repository ID and a connection handle for further
// communication.
func DialWebSocket(ctx context.Context, u string, id repo.RepoID) (*WSConn, repo.RepoID, error) {
	// ensure scheme is ws/wss
	parsed, err := url.Parse(u)
	if err != nil {
		return nil, repo.RepoID{}, err
	}
	if parsed.Scheme != "ws" && parsed.Scheme != "wss" {
		return nil, repo.RepoID{}, fmt.Errorf("invalid websocket url: %s", u)
	}
	dialer := websocket.DefaultDialer
	conn, _, err := dialer.DialContext(ctx, u, nil)
	if err != nil {
		return nil, repo.RepoID{}, err
	}
	ws := NewWSConn(conn)
	if d, ok := ctx.Deadline(); ok {
		_ = conn.SetReadDeadline(d)
		_ = conn.SetWriteDeadline(d)
		defer conn.SetReadDeadline(time.Time{})
		defer conn.SetWriteDeadline(time.Time{})
	}
	if err := ws.Send(handshakeMessage{Type: "join", SenderID: id.String()}); err != nil {
		ws.Close()
		return nil, repo.RepoID{}, err
	}
	var resp handshakeMessage
	if err := ws.Recv(&resp); err != nil {
		ws.Close()
		return nil, repo.RepoID{}, err
	}
	if resp.Type != "peer" {
		ws.Close()
		return nil, repo.RepoID{}, fmt.Errorf("unexpected message %q", resp.Type)
	}
	remote := parseRepoID(resp.SenderID)
	return ws, remote, nil
}

// AcceptWebSocket upgrades an HTTP request to a websocket and completes the
// join/peer handshake. The returned connection can be used for CBOR message
// exchange.
func AcceptWebSocket(w http.ResponseWriter, r *http.Request, id repo.RepoID) (*WSConn, repo.RepoID, error) {
	upgrader := websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, repo.RepoID{}, err
	}
	ws := NewWSConn(conn)
	var req handshakeMessage
	if err := ws.Recv(&req); err != nil {
		ws.Close()
		return nil, repo.RepoID{}, err
	}
	if req.Type != "join" {
		ws.Close()
		return nil, repo.RepoID{}, fmt.Errorf("unexpected message %q", req.Type)
	}
	if err := ws.Send(handshakeMessage{Type: "peer", SenderID: id.String()}); err != nil {
		ws.Close()
		return nil, repo.RepoID{}, err
	}
	remote := parseRepoID(req.SenderID)
	return ws, remote, nil
}

// parseRepoID attempts to parse the given string as a UUID. If parsing fails
// (for example when the sender uses non-UUID identifiers like "peer-rgp224jx")
// it deterministically generates a UUID v5 (SHA-1) from the original string.
//
// This allows interop with JS implementations that use human-readable peer IDs
// while keeping the rest of the codebase working with the uuid.UUID type.
func parseRepoID(s string) uuid.UUID {
	if id, err := uuid.Parse(s); err == nil {
		return id
	}
	// Use the string to derive a deterministic UUID so that the same peer ID
	// maps to the same UUID across connections.
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(s))
}