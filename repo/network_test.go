package repo

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestHandshake(t *testing.T) {
	c1, c2 := net.Pipe()
	repo1 := New()
	repo2 := New()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r1, r2, err := handshakePipe(ctx, c1, Outgoing, repo1.ID, c2, Incoming, repo2.ID)
	if err != nil {
		t.Fatalf("handshake error: %v", err)
	}
	if r1 != repo2.ID || r2 != repo1.ID {
		t.Fatalf("unexpected IDs: %v %v", r1, r2)
	}
}
