package repo

import (
	"fmt"
	"sync"

	automerge "github.com/automerge/automerge-go"
	"github.com/google/uuid"
)

// DocumentID uniquely identifies a document.
type DocumentID = uuid.UUID

// RepoID uniquely identifies a repository instance.
type RepoID = uuid.UUID

// Document represents a single Automerge document.
type Document struct {
	ID  DocumentID
	doc *automerge.Doc

	watchers   []chan struct{}
	watchersMu sync.Mutex
}

// NewSyncState returns a sync state for exchanging changes of this document with a peer.
func (d *Document) NewSyncState() *automerge.SyncState {
	if d.doc == nil {
		d.doc = automerge.New()
	}
	return automerge.NewSyncState(d.doc)
}

// ReceiveSyncMessage applies a sync message to the document using the given state.
func (d *Document) ReceiveSyncMessage(state *automerge.SyncState, msg []byte) error {
	state.Doc = d.doc
	_, err := state.ReceiveMessage(msg)
	if err == nil {
		d.notifyWatchers()
	}
	return err
}

// GenerateSyncMessage produces the next sync message for the peer using the given state.
func (d *Document) GenerateSyncMessage(state *automerge.SyncState) ([]byte, bool) {
	state.Doc = d.doc
	sm, valid := state.GenerateMessage()
	if !valid {
		return nil, false
	}
	return sm.Bytes(), true
}

// Set assigns a value in the document.
func (d *Document) Set(key string, value interface{}) error {
	if d.doc == nil {
		d.doc = automerge.New()
	}
	if err := d.doc.RootMap().Set(key, value); err != nil {
		return err
	}
	_, err := d.doc.Commit("set")
	if err == nil {
		d.notifyWatchers()
	}
	return err
}

// Get retrieves a value from the document.
func (d *Document) Get(key string) (interface{}, bool) {
	if d.doc == nil {
		return nil, false
	}
	v, err := automerge.As[interface{}](d.doc.RootMap().Get(key))
	if err != nil {
		return nil, false
	}
	return v, true
}

// Repo holds a collection of documents.
type Repo struct {
	ID    RepoID
	docs  map[DocumentID]*Document
	store *FsStore
}

// New returns a new empty repository with a random identifier.
func New() *Repo {
	return &Repo{ID: uuid.New(), docs: make(map[DocumentID]*Document)}
}

// NewWithStore creates a repository that will persist documents using the provided store.
func NewWithStore(store *FsStore) *Repo {
	r := New()
	r.store = store
	return r
}

// NewDoc creates a new document within the repository and returns it.
func (r *Repo) NewDoc() *Document {
	doc := &Document{ID: uuid.New(), doc: automerge.New()}
	r.docs[doc.ID] = doc
	return doc
}

// GetDoc retrieves a document by id.
func (r *Repo) GetDoc(id DocumentID) (*Document, bool) {
	d, ok := r.docs[id]
	return d, ok
}

// SaveDoc writes a document to disk using the repo's store.
func (r *Repo) SaveDoc(id DocumentID) error {
	if r.store == nil {
		return fmt.Errorf("no store configured")
	}
	doc, ok := r.docs[id]
	if !ok {
		return fmt.Errorf("document %s not found", id)
	}
	return r.store.Save(doc)
}

// LoadDoc loads a document from disk into the repo.
func (r *Repo) LoadDoc(id DocumentID) (*Document, error) {
	if r.store == nil {
		return nil, fmt.Errorf("no store configured")
	}
	doc, err := r.store.Load(id)
	if err != nil {
		return nil, err
	}
	r.docs[id] = doc
	return doc, nil
}
