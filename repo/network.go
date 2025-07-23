package repo

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

// ConnDirection indicates if we initiated a connection or accepted one.
type ConnDirection int

const (
	Incoming ConnDirection = iota
	Outgoing
)

type handshakeMessage struct {
	Type     string `json:"type"`
	SenderID RepoID `json:"senderId"`
}

// Handshake performs a simple join/peer handshake over the given connection.
// It returns the remote repository ID after the handshake completes.
func Handshake(rw io.ReadWriter, id RepoID, dir ConnDirection) (RepoID, error) {
	enc := json.NewEncoder(rw)
	dec := json.NewDecoder(bufio.NewReader(rw))
	switch dir {
	case Outgoing:
		if err := enc.Encode(handshakeMessage{Type: "join", SenderID: id}); err != nil {
			return RepoID{}, err
		}
		var resp handshakeMessage
		if err := dec.Decode(&resp); err != nil {
			return RepoID{}, err
		}
		if resp.Type != "peer" {
			return RepoID{}, fmt.Errorf("unexpected message %q", resp.Type)
		}
		return resp.SenderID, nil
	case Incoming:
		var req handshakeMessage
		if err := dec.Decode(&req); err != nil {
			return RepoID{}, err
		}
		if req.Type != "join" {
			return RepoID{}, fmt.Errorf("unexpected message %q", req.Type)
		}
		if err := enc.Encode(handshakeMessage{Type: "peer", SenderID: id}); err != nil {
			return RepoID{}, err
		}
		return req.SenderID, nil
	default:
		return RepoID{}, fmt.Errorf("invalid direction")
	}
}

// handshakePipe is a helper for tests that connects two sides of a net.Pipe and
// runs Handshake concurrently.
func handshakePipe(c1 io.ReadWriter, dir1 ConnDirection, id1 RepoID, c2 io.ReadWriter, dir2 ConnDirection, id2 RepoID) (RepoID, RepoID, error) {
	var wg sync.WaitGroup
	var r1 RepoID
	var e1 error
	var r2 RepoID
	var e2 error
	wg.Add(2)
	go func() {
		defer wg.Done()
		r1, e1 = Handshake(c1, id1, dir1)
	}()
	go func() {
		defer wg.Done()
		r2, e2 = Handshake(c2, id2, dir2)
	}()
	wg.Wait()
	if e1 != nil {
		return RepoID{}, RepoID{}, e1
	}
	if e2 != nil {
		return RepoID{}, RepoID{}, e2
	}
	return r1, r2, nil
}
