package repo

import (
	"net"
	"testing"
)

func TestHandshake(t *testing.T) {
	c1, c2 := net.Pipe()
	repo1 := New()
	repo2 := New()

	r1, r2, err := handshakePipe(c1, Outgoing, repo1.ID, c2, Incoming, repo2.ID)
	if err != nil {
		t.Fatalf("handshake error: %v", err)
	}
	if r1 != repo2.ID || r2 != repo1.ID {
		t.Fatalf("unexpected IDs: %v %v", r1, r2)
	}
}
