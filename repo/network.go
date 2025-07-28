package repo

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/fxamacker/cbor/v2"
)

// ConnDirection indicates if we initiated a connection or accepted one.
type ConnDirection int

const (
	Incoming ConnDirection = iota
	Outgoing
)


type handshakeMessage struct {
	Type     string `cbor:"type"`
	SenderID string `cbor:"senderId"`
}

// Handshake performs a simple join/peer handshake over the given connection.
// It returns the remote repository ID after the handshake completes.
func Handshake(ctx context.Context, rw io.ReadWriter, id RepoID, dir ConnDirection) (RepoID, error) {
	if conn, ok := rw.(interface{ SetDeadline(time.Time) error }); ok {
		if d, ok := ctx.Deadline(); ok {
			_ = conn.SetDeadline(d)
			defer conn.SetDeadline(time.Time{})
		}
	}

	enc := cbor.NewEncoder(rw)
	dec := cbor.NewDecoder(bufio.NewReader(rw))
	switch dir {
	case Outgoing:
		if err := enc.Encode(handshakeMessage{Type: "join", SenderID: id.String()}); err != nil {
			return RepoID{}, err
		}
		var resp handshakeMessage
		if err := dec.Decode(&resp); err != nil {
			return RepoID{}, err
		}
		if resp.Type != "peer" {
			return RepoID{}, fmt.Errorf("unexpected message %q", resp.Type)
		}
		log.Printf("The user is sending UUID: %s", resp.SenderID)
		// Accept arbitrary peer identifiers. If it's not a UUID we map it to a
		// deterministic UUID so the rest of the code can continue to use the
		// uuid.UUID type.
		remote := parseRepoID(resp.SenderID)
		return remote, nil
	case Incoming:
		var req handshakeMessage
		if err := dec.Decode(&req); err != nil {
			return RepoID{}, err
		}
		if req.Type != "join" {
			return RepoID{}, fmt.Errorf("unexpected message %q", req.Type)
		}
		log.Printf("The user is sending UUID: %s", req.SenderID)
		if err := enc.Encode(handshakeMessage{Type: "peer", SenderID: id.String()}); err != nil {
			return RepoID{}, err
		}
		remote := parseRepoID(req.SenderID)
		return remote, nil
	default:
		return RepoID{}, fmt.Errorf("invalid direction")
	}
}


// handshakePipe is a helper for tests that connects two sides of a net.Pipe and
// runs Handshake concurrently.
func handshakePipe(ctx context.Context, c1 io.ReadWriter, dir1 ConnDirection, id1 RepoID, c2 io.ReadWriter, dir2 ConnDirection, id2 RepoID) (RepoID, RepoID, error) {
	var wg sync.WaitGroup
	var r1 RepoID
	var e1 error
	var r2 RepoID
	var e2 error
	wg.Add(2)
	go func() {
		defer wg.Done()
		r1, e1 = Handshake(ctx, c1, id1, dir1)
	}()
	go func() {
		defer wg.Done()
		r2, e2 = Handshake(ctx, c2, id2, dir2)
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
