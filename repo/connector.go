package repo

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/fxamacker/cbor/v2"
)

// LPConn wraps a connection and exchanges length-prefixed CBOR messages.
type LPConn struct {
	rw io.ReadWriteCloser
	mu sync.Mutex
}

// NewLPConn returns a new length prefixed connection.
func NewLPConn(rw io.ReadWriteCloser) *LPConn {
	return &LPConn{rw: rw}
}

// Send encodes v as CBOR and writes it with a 4 byte length prefix.
func (c *LPConn) Send(v interface{}) error {
	data, err := cbor.Marshal(v)
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

// Recv reads a length prefixed CBOR message into v.
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
	return cbor.Unmarshal(data, v)
}

// SendMessage writes a RepoMessage using length-prefixed CBOR encoding.
func (c *LPConn) SendMessage(msg RepoMessage) error {
	data, err := msg.Encode()
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

// RecvMessage reads a RepoMessage that was sent using SendMessage.
func (c *LPConn) RecvMessage() (RepoMessage, error) {
	var lenBuf [4]byte
	if _, err := io.ReadFull(c.rw, lenBuf[:]); err != nil {
		return RepoMessage{}, err
	}
	n := binary.BigEndian.Uint32(lenBuf[:])
	data := make([]byte, n)
	if _, err := io.ReadFull(c.rw, data); err != nil {
		return RepoMessage{}, err
	}
	return DecodeRepoMessage(data)
}

// Close closes the underlying connection.
func (c *LPConn) Close() error { return c.rw.Close() }

// Connect performs a handshake over conn using length-prefixed messages and
// returns the remote repo ID along with a LPConn for further communication.
func Connect(ctx context.Context, conn net.Conn, id RepoID, dir ConnDirection) (*LPConn, RepoID, error) {
	if d, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(d)
		defer conn.SetDeadline(time.Time{})
	}

	lp := NewLPConn(conn)
	switch dir {
	case Outgoing:
		if err := lp.Send(handshakeMessage{Type: "join", SenderID: id.String()}); err != nil {
			return nil, RepoID{}, err
		}
		var resp handshakeMessage
		if err := lp.Recv(&resp); err != nil {
			return nil, RepoID{}, err
		}
		if resp.Type != "peer" {
			return nil, RepoID{}, fmt.Errorf("unexpected message %q", resp.Type)
		}
		remote := parseRepoID(resp.SenderID)
		return lp, remote, nil
	case Incoming:
		var req handshakeMessage
		if err := lp.Recv(&req); err != nil {
			return nil, RepoID{}, err
		}
		if req.Type != "join" {
			return nil, RepoID{}, fmt.Errorf("unexpected message %q", req.Type)
		}
		if err := lp.Send(handshakeMessage{Type: "peer", SenderID: id.String()}); err != nil {
			return nil, RepoID{}, err
		}
		remote := parseRepoID(req.SenderID)
		return lp, remote, nil
	default:
		return nil, RepoID{}, fmt.Errorf("invalid direction")
	}
}