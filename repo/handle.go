package repo

import (
	"fmt"
	"sync"

	automerge "github.com/automerge/automerge-go"
)

// HandleEvent represents a peer connection lifecycle event emitted by RepoHandle.
type HandleEvent struct {
	Type string
	Peer RepoID
	Err  error
}

const (
	EventPeerConnected    = "peer_connected"
	EventPeerDisconnected = "peer_disconnected"
	EventConnError        = "conn_error"
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
	peers map[RepoID]*peerInfo

	// Inbox delivers messages received from peers. It is unbuffered so callers
	// should read from it promptly.
	Inbox chan RepoMessage

	// Events publishes connection lifecycle notifications such as when peers
	// connect or disconnect.
	Events chan HandleEvent
}

func (h *RepoHandle) emitEvent(e HandleEvent) {
	if h.Events == nil {
		return
	}
	defer func() { recover() }()
	h.Events <- e
}

type peerInfo struct {
	conn       Conn
	syncStates map[DocumentID]*automerge.SyncState
}

// NewRepoHandle wraps r with connection management and returns the handle.
func NewRepoHandle(r *Repo) *RepoHandle {
	return &RepoHandle{
		Repo:   r,
		peers:  make(map[RepoID]*peerInfo),
		Inbox:  make(chan RepoMessage),
		Events: make(chan HandleEvent, 8),
	}
}

// AddConn registers a connection to a remote peer and starts a goroutine to
// forward its messages onto the handle's Inbox channel.
func (h *RepoHandle) AddConn(remote RepoID, c Conn) {
	h.mu.Lock()
	if h.peers == nil {
		h.peers = make(map[RepoID]*peerInfo)
	}
	h.peers[remote] = &peerInfo{conn: c, syncStates: make(map[DocumentID]*automerge.SyncState)}
	h.mu.Unlock()

	go h.readLoop(remote, c)
	h.emitEvent(HandleEvent{Type: EventPeerConnected, Peer: remote})
}

// readLoop continuously receives messages from c and publishes them to Inbox.
func (h *RepoHandle) readLoop(remote RepoID, c Conn) {
	for {
		msg, err := c.RecvMessage()
		if err != nil {
			h.emitEvent(HandleEvent{Type: EventConnError, Peer: remote, Err: err})
			break
		}
		if msg.Type == "sync" {
			h.handleSyncMessage(remote, msg)
			continue
		}
		h.Inbox <- msg
	}
	h.RemoveConn(remote)
}

// RemoveConn closes and deletes the connection associated with the peer.
func (h *RepoHandle) RemoveConn(remote RepoID) {
	h.mu.Lock()
	pi, ok := h.peers[remote]
	if ok {
		delete(h.peers, remote)
	}
	h.mu.Unlock()

	if ok {
		pi.conn.Close()
		h.emitEvent(HandleEvent{Type: EventPeerDisconnected, Peer: remote})
	}
}

// SendMessage transmits msg to the specified remote peer if present.
func (h *RepoHandle) SendMessage(remote RepoID, msg RepoMessage) error {
	h.mu.Lock()
	pi, ok := h.peers[remote]
	h.mu.Unlock()
	if !ok {
		return fmt.Errorf("peer %s not found", remote)
	}
	if err := pi.conn.SendMessage(msg); err != nil {
		h.emitEvent(HandleEvent{Type: EventConnError, Peer: remote, Err: err})
		h.RemoveConn(remote)
		return err
	}
	return nil
}

// Broadcast sends msg to all connected peers. Errors are returned for the first
// failure encountered.
func (h *RepoHandle) Broadcast(msg RepoMessage) error {
	h.mu.Lock()
	conns := make([]Conn, 0, len(h.peers))
	ids := make([]RepoID, 0, len(h.peers))
	for id := range h.peers {
		ids = append(ids, id)
		conns = append(conns, h.peers[id].conn)
	}
	h.mu.Unlock()
	for i, c := range conns {
		if err := c.SendMessage(msg); err != nil {
			remote := ids[i]
			h.emitEvent(HandleEvent{Type: EventConnError, Peer: remote, Err: err})
			h.RemoveConn(remote)
			return err
		}
	}
	return nil
}

// Close terminates all peer connections and closes the Inbox channel.
func (h *RepoHandle) Close() {
	h.mu.Lock()
	conns := h.peers
	h.peers = make(map[RepoID]*peerInfo)
	h.mu.Unlock()
	for _, pi := range conns {
		pi.conn.Close()
	}
	close(h.Inbox)
	if h.Events != nil {
		close(h.Events)
	}
}

// SyncDocument exchanges sync messages for the given document with the remote peer.
func (h *RepoHandle) SyncDocument(remote RepoID, docID DocumentID) error {
	h.mu.Lock()
	pi, ok := h.peers[remote]
	doc, docOK := h.Repo.GetDoc(docID)
	if ok && docOK {
		state := pi.syncStates[docID]
		if state == nil {
			state = doc.NewSyncState()
			pi.syncStates[docID] = state
		}
		h.mu.Unlock()
		for {
			data, valid := doc.GenerateSyncMessage(state)
			if !valid {
				break
			}
			msg := RepoMessage{Type: "sync", FromRepoID: h.Repo.ID, ToRepoID: remote, DocumentID: docID, Message: data}
			if err := pi.conn.SendMessage(msg); err != nil {
				return err
			}
		}
		return nil
	}
	h.mu.Unlock()
	if !ok {
		return fmt.Errorf("peer %s not found", remote)
	}
	return fmt.Errorf("document %s not found", docID)
}

// handleSyncMessage applies a sync message from a peer and responds with any updates.
func (h *RepoHandle) handleSyncMessage(remote RepoID, msg RepoMessage) {
	h.mu.Lock()
	pi, ok := h.peers[remote]
	if !ok {
		h.mu.Unlock()
		return
	}
	doc, docOK := h.Repo.GetDoc(msg.DocumentID)
	if !docOK {
		// create empty document if not present
		doc = &Document{ID: msg.DocumentID, doc: automerge.New()}
		h.Repo.docs[msg.DocumentID] = doc
	}
	state := pi.syncStates[msg.DocumentID]
	if state == nil {
		state = doc.NewSyncState()
		pi.syncStates[msg.DocumentID] = state
	}
	h.mu.Unlock()

	_ = doc.ReceiveSyncMessage(state, msg.Message)
	_ = h.SyncDocument(remote, msg.DocumentID)
}

// SyncAll sends sync messages for all documents to the remote peer.
func (h *RepoHandle) SyncAll(remote RepoID) error {
	h.mu.Lock()
	ids := make([]DocumentID, 0, len(h.Repo.docs))
	for id := range h.Repo.docs {
		ids = append(ids, id)
	}
	h.mu.Unlock()
	for _, id := range ids {
		if err := h.SyncDocument(remote, id); err != nil {
			return err
		}
	}
	return nil
}
