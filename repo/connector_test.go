package repo

import (
	"net"
	"testing"
)

func TestConnect(t *testing.T) {
	c1, c2 := net.Pipe()
	repo1 := New()
	repo2 := New()

	var remote1 RepoID
	var remote2 RepoID
	var lp1 *LPConn
	var lp2 *LPConn
	errCh := make(chan error, 2)
	go func() {
		var err error
		lp1, remote1, err = Connect(c1, repo1.ID, Outgoing)
		errCh <- err
	}()
	go func() {
		var err error
		lp2, remote2, err = Connect(c2, repo2.ID, Incoming)
		errCh <- err
	}()
	if err := <-errCh; err != nil {
		t.Fatalf("connect error: %v", err)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("connect error: %v", err)
	}
	if remote1 != repo2.ID || remote2 != repo1.ID {
		t.Fatalf("unexpected repo IDs: %v %v", remote1, remote2)
	}
	lp1.Close()
	lp2.Close()
}
