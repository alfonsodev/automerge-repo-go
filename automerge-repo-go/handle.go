package repo

import (
	"context"
	"fmt"
	"sync"
	"time"

	automerge "github.com/automerge/automerge-go"
)

// HandleEvent represents a peer connection lifecycle event emitted by RepoHandle.
type HandleEvent struct {
	Type string
	Peer RepoID
	Err  error
}

const (
	// EventPeerConnected is emitted when a new peer is connected.
	EventPeerConnected = "peer_connected"
	// EventPeerDisconnected is emitted when a peer disconnects.
	EventPeerDisconnected = "peer_disconnected"
	// EventConnError is emitted when a connection error occurs.
	EventConnError = "conn_error"
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
// onto the Inbox channel. It is the primary way to interact with the network.
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

// ConnFinishedKind describes why a connection goroutine exited.
type ConnFinishedKind int

const (
	// ConnFinishedRecvError indicates the connection ended because a receive operation failed.
	ConnFinishedRecvError ConnFinishedKind = iota
	// ConnFinishedSendError indicates the connection ended because a send operation failed.
	ConnFinishedSendError
	// ConnFinishedLocalClose indicates the connection was closed locally via RemoveConn.
	ConnFinishedLocalClose
)

// ConnFinished provides the reason a connection goroutine exited.
type ConnFinished struct {
	Kind ConnFinishedKind
	Err  error
}

// ConnComplete is returned by AddConn and resolves when the connection
// goroutine exits. The ConnFinished value indicates why the connection finished.
type ConnComplete struct{ ch <-chan ConnFinished }

// Await blocks until the connection has completed and returns the reason.
func (c ConnComplete) Await() ConnFinished {
	return <-c.ch
}

type peerInfo struct {
	conn       Conn
	complete   chan ConnFinished
	syncStates map[DocumentID]*automerge.SyncState
}

// NewRepoHandle wraps r with connection management and returns the handle.
func NewRepoHandle(r *Repo) *RepoHandle {
	return &RepoHandle{
		Repo:   r,
		peers:  make(map[RepoID]*peerInfo),
		Inbox:  make(chan RepoMessage, 16),
		Events: make(chan HandleEvent, 8),
	}
}

// AddConn registers a connection to a remote peer and starts a goroutine to
// forward its messages onto the handle's Inbox channel. It returns a
// ConnComplete that resolves when the connection goroutine exits.
func (h *RepoHandle) AddConn(remote RepoID, c Conn) ConnComplete {
	h.mu.Lock()
	if h.peers == nil {
		h.peers = make(map[RepoID]*peerInfo)
	}
	done := make(chan ConnFinished, 1)
	h.peers[remote] = &peerInfo{conn: c, complete: done, syncStates: make(map[DocumentID]*automerge.SyncState)}
	h.mu.Unlock()

	go h.readLoop(remote, c, done)
	h.emitEvent(HandleEvent{Type: EventPeerConnected, Peer: remote})
	return ConnComplete{ch: done}
}

// AddConnWithRetry repeatedly dials the remote using dial and registers the
// connection with AddConn. If the connection closes with an error it will be
// retried after delay until ctx is canceled. The returned ConnComplete resolves
// when the retry loop exits.
func (h *RepoHandle) AddConnWithRetry(ctx context.Context, remote RepoID, dial func(context.Context) (Conn, error), delay time.Duration) ConnComplete {
	done := make(chan ConnFinished, 1)
	go func() {
		for {
			select {
			case <-ctx.Done():
				h.removePeer(remote, ConnFinished{Kind: ConnFinishedLocalClose, Err: ctx.Err()})
				done <- ConnFinished{Kind: ConnFinishedLocalClose, Err: ctx.Err()}
				close(done)
				return
			default:
			}

			conn, err := dial(ctx)
			if err != nil {
				done <- ConnFinished{Kind: ConnFinishedRecvError, Err: err}
				close(done)
				return
			}
			cc := h.AddConn(remote, conn)
			cc.Await()

			if ctx.Err() != nil {
				h.removePeer(remote, ConnFinished{Kind: ConnFinishedLocalClose, Err: ctx.Err()})
				done <- ConnFinished{Kind: ConnFinishedLocalClose, Err: ctx.Err()}
				close(done)
				return
			}

			// wait before reconnecting regardless of reason
			time.Sleep(delay)
			continue
		}
	}()
	return ConnComplete{ch: done}
}

// readLoop continuously receives messages from c and publishes them to Inbox.
func (h *RepoHandle) readLoop(remote RepoID, c Conn, done chan ConnFinished) {
	var err error
	for {
		var msg RepoMessage
		msg, err = c.RecvMessage()
		if err != nil {
			h.emitEvent(HandleEvent{Type: EventConnError, Peer: remote, Err: err})
			break
		}
		if msg.Type == "sync" {
			h.handleSyncMessage(remote, msg)
			continue
		}
		fmt.Printf("readLoop: Sending message type %s to Inbox for doc %s\n", msg.Type, msg.DocumentID)
		h.Inbox <- msg
	}
	h.removePeer(remote, ConnFinished{Kind: ConnFinishedRecvError, Err: err})
}

// RemoveConn closes and deletes the connection associated with the peer.
func (h *RepoHandle) RemoveConn(remote RepoID) {
	h.removePeer(remote, ConnFinished{Kind: ConnFinishedLocalClose})
}

func (h *RepoHandle) removePeer(remote RepoID, reason ConnFinished) {
	h.mu.Lock()
	pi, ok := h.peers[remote]
	if ok {
		delete(h.peers, remote)
	}
	h.mu.Unlock()

	if ok {
		pi.conn.Close()
		h.emitEvent(HandleEvent{Type: EventPeerDisconnected, Peer: remote})
		if pi.complete != nil {
			pi.complete <- reason
			close(pi.complete)
		}
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
		h.removePeer(remote, ConnFinished{Kind: ConnFinishedSendError, Err: err})
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
			h.removePeer(remote, ConnFinished{Kind: ConnFinishedSendError, Err: err})
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
	for id, pi := range conns {
		pi.conn.Close()
		if pi.complete != nil {
			pi.complete <- ConnFinished{Kind: ConnFinishedLocalClose}
			close(pi.complete)
		}
		h.emitEvent(HandleEvent{Type: EventPeerDisconnected, Peer: id})
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
		if h.Repo.sharePolicy != nil && h.Repo.sharePolicy.ShouldSync(docID, remote) == DontShare {
			h.mu.Unlock()
			return nil
		}
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
	if h.Repo.sharePolicy != nil && h.Repo.sharePolicy.ShouldSync(msg.DocumentID, remote) == DontShare {
		h.mu.Unlock()
		return
	}
	doc, docOK := h.Repo.GetDoc(msg.DocumentID)
	if !docOK {
		if h.Repo.sharePolicy != nil && h.Repo.sharePolicy.ShouldRequest(msg.DocumentID, remote) == DontShare {
			h.mu.Unlock()
			return
		}
		// create empty document if not present
		doc = &Document{ID: msg.DocumentID, Doc: automerge.New()}
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
		if h.Repo.sharePolicy != nil && h.Repo.sharePolicy.ShouldAnnounce(id, remote) == DontShare {
			continue
		}
		if err := h.SyncDocument(remote, id); err != nil {
			return err
		}
	}
	return nil
}