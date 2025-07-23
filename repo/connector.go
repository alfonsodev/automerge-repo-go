package repo

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
)

// LPConn wraps a connection and exchanges length-prefixed JSON messages.
type LPConn struct {
	rw io.ReadWriteCloser
	mu sync.Mutex
}

// NewLPConn returns a new length prefixed connection.
func NewLPConn(rw io.ReadWriteCloser) *LPConn {
	return &LPConn{rw: rw}
}

// Send encodes v as JSON and writes it with a 4 byte length prefix.
func (c *LPConn) Send(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	var lenBuf [4]byte
	binary.BigEndian.PutUint32(lenBuf[:], uint32(len(data)))
	if _, err := c.rw.Write(lenBuf[:]); err != nil {
		return err
	}
	_, err = c.rw.Write(data)
	return err
}

// Recv reads a length prefixed JSON message into v.
func (c *LPConn) Recv(v interface{}) error {
	var lenBuf [4]byte
	if _, err := io.ReadFull(c.rw, lenBuf[:]); err != nil {
		return err
	}
	n := binary.BigEndian.Uint32(lenBuf[:])
	data := make([]byte, n)
	if _, err := io.ReadFull(c.rw, data); err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// Close closes the underlying connection.
func (c *LPConn) Close() error { return c.rw.Close() }

// Connect performs a handshake over conn using length-prefixed messages and
// returns the remote repo ID along with a LPConn for further communication.
func Connect(conn net.Conn, id RepoID, dir ConnDirection) (*LPConn, RepoID, error) {
	lp := NewLPConn(conn)
	switch dir {
	case Outgoing:
		if err := lp.Send(handshakeMessage{Type: "join", SenderID: id}); err != nil {
			return nil, RepoID{}, err
		}
		var resp handshakeMessage
		if err := lp.Recv(&resp); err != nil {
			return nil, RepoID{}, err
		}
		if resp.Type != "peer" {
			return nil, RepoID{}, fmt.Errorf("unexpected message %q", resp.Type)
		}
		return lp, resp.SenderID, nil
	case Incoming:
		var req handshakeMessage
		if err := lp.Recv(&req); err != nil {
			return nil, RepoID{}, err
		}
		if req.Type != "join" {
			return nil, RepoID{}, fmt.Errorf("unexpected message %q", req.Type)
		}
		if err := lp.Send(handshakeMessage{Type: "peer", SenderID: id}); err != nil {
			return nil, RepoID{}, err
		}
		return lp, req.SenderID, nil
	default:
		return nil, RepoID{}, fmt.Errorf("invalid direction")
	}
}
