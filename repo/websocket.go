package repo

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/gorilla/websocket"
)

// WSConn wraps a websocket connection for sending JSON messages.
type WSConn struct {
	c  *websocket.Conn
	mu sync.Mutex
}

// NewWSConn creates a new WSConn.
func NewWSConn(c *websocket.Conn) *WSConn {
	return &WSConn{c: c}
}

// Send encodes v as JSON and writes it over the websocket.
func (c *WSConn) Send(v interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.WriteJSON(v)
}

// Recv reads a JSON message into v.
func (c *WSConn) Recv(v interface{}) error {
	return c.c.ReadJSON(v)
}

// Close closes the websocket.
func (c *WSConn) Close() error { return c.c.Close() }

// DialWebSocket dials the given websocket URL and performs the join/peer handshake.
// It returns the remote repository ID and a connection handle for further
// communication.
func DialWebSocket(u string, id RepoID) (*WSConn, RepoID, error) {
	// ensure scheme is ws/wss
	parsed, err := url.Parse(u)
	if err != nil {
		return nil, RepoID{}, err
	}
	if parsed.Scheme != "ws" && parsed.Scheme != "wss" {
		return nil, RepoID{}, fmt.Errorf("invalid websocket url: %s", u)
	}
	conn, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		return nil, RepoID{}, err
	}
	ws := NewWSConn(conn)
	if err := ws.Send(handshakeMessage{Type: "join", SenderID: id}); err != nil {
		ws.Close()
		return nil, RepoID{}, err
	}
	var resp handshakeMessage
	if err := ws.Recv(&resp); err != nil {
		ws.Close()
		return nil, RepoID{}, err
	}
	if resp.Type != "peer" {
		ws.Close()
		return nil, RepoID{}, fmt.Errorf("unexpected message %q", resp.Type)
	}
	return ws, resp.SenderID, nil
}

// AcceptWebSocket upgrades an HTTP request to a websocket and completes the
// join/peer handshake. The returned connection can be used for JSON message
// exchange.
func AcceptWebSocket(w http.ResponseWriter, r *http.Request, id RepoID) (*WSConn, RepoID, error) {
	upgrader := websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, RepoID{}, err
	}
	ws := NewWSConn(conn)
	var req handshakeMessage
	if err := ws.Recv(&req); err != nil {
		ws.Close()
		return nil, RepoID{}, err
	}
	if req.Type != "join" {
		ws.Close()
		return nil, RepoID{}, fmt.Errorf("unexpected message %q", req.Type)
	}
	if err := ws.Send(handshakeMessage{Type: "peer", SenderID: id}); err != nil {
		ws.Close()
		return nil, RepoID{}, err
	}
	return ws, req.SenderID, nil
}
