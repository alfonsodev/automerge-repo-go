package repo

import (
	"fmt"
	"sync"
)

// Conn abstracts a bidirectional channel capable of sending and receiving
// RepoMessage values. LPConn and WSConn satisfy this interface.
type Conn interface {
	SendMessage(RepoMessage) error
	RecvMessage() (RepoMessage, error)
	Close() error
}

// RepoHandle manages a Repo along with its active peer connections. It
// spawns goroutines for each connection that forward incoming messages
// onto the Inbox channel.
type RepoHandle struct {
	Repo *Repo

	mu    sync.Mutex
	peers map[RepoID]Conn

	// Inbox delivers messages received from peers. It is unbuffered so callers
	// should read from it promptly.
	Inbox chan RepoMessage
}

// NewRepoHandle wraps r with connection management and returns the handle.
func NewRepoHandle(r *Repo) *RepoHandle {
	return &RepoHandle{
		Repo:  r,
		peers: make(map[RepoID]Conn),
		Inbox: make(chan RepoMessage),
	}
}

// AddConn registers a connection to a remote peer and starts a goroutine to
// forward its messages onto the handle's Inbox channel.
func (h *RepoHandle) AddConn(remote RepoID, c Conn) {
	h.mu.Lock()
	if h.peers == nil {
		h.peers = make(map[RepoID]Conn)
	}
	h.peers[remote] = c
	h.mu.Unlock()

	go h.readLoop(remote, c)
}

// readLoop continuously receives messages from c and publishes them to Inbox.
func (h *RepoHandle) readLoop(remote RepoID, c Conn) {
	for {
		msg, err := c.RecvMessage()
		if err != nil {
			break
		}
		h.Inbox <- msg
	}
	h.RemoveConn(remote)
}

// RemoveConn closes and deletes the connection associated with the peer.
func (h *RepoHandle) RemoveConn(remote RepoID) {
	h.mu.Lock()
	c, ok := h.peers[remote]
	if ok {
		delete(h.peers, remote)
	}
	h.mu.Unlock()

	if ok {
		c.Close()
	}
}

// SendMessage transmits msg to the specified remote peer if present.
func (h *RepoHandle) SendMessage(remote RepoID, msg RepoMessage) error {
	h.mu.Lock()
	c, ok := h.peers[remote]
	h.mu.Unlock()
	if !ok {
		return fmt.Errorf("peer %s not found", remote)
	}
	return c.SendMessage(msg)
}

// Broadcast sends msg to all connected peers. Errors are returned for the first
// failure encountered.
func (h *RepoHandle) Broadcast(msg RepoMessage) error {
	h.mu.Lock()
	conns := make([]Conn, 0, len(h.peers))
	for _, c := range h.peers {
		conns = append(conns, c)
	}
	h.mu.Unlock()
	for _, c := range conns {
		if err := c.SendMessage(msg); err != nil {
			return err
		}
	}
	return nil
}

// Close terminates all peer connections and closes the Inbox channel.
func (h *RepoHandle) Close() {
	h.mu.Lock()
	conns := h.peers
	h.peers = make(map[RepoID]Conn)
	h.mu.Unlock()
	for _, c := range conns {
		c.Close()
	}
	close(h.Inbox)
}
