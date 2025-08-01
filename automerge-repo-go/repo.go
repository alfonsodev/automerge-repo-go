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
	Doc *automerge.Doc

	lastHeads           []automerge.ChangeHash
	changesSinceCompact int

	watchers   []chan struct{}
	watchersMu sync.Mutex
}

// NewSyncState returns a sync state for exchanging changes of this document with a peer.
func (d *Document) NewSyncState() *automerge.SyncState {
	if d.Doc == nil {
		d.Doc = automerge.New()
	}
	return automerge.NewSyncState(d.Doc)
}

// ReceiveSyncMessage applies a sync message to the document using the given state.
func (d *Document) ReceiveSyncMessage(state *automerge.SyncState, msg []byte) error {
	state.Doc = d.Doc
	_, err := state.ReceiveMessage(msg)
	if err == nil {
		d.notifyWatchers()
	}
	return err
}

// GenerateSyncMessage produces the next sync message for the peer using the given state.
func (d *Document) GenerateSyncMessage(state *automerge.SyncState) ([]byte, bool) {
	state.Doc = d.Doc
	sm, valid := state.GenerateMessage()
	if !valid {
		return nil, false
	}
	return sm.Bytes(), true
}

// Set assigns a value in the document.
func (d *Document) Set(key string, value interface{}) error {
	if d.Doc == nil {
		d.Doc = automerge.New()
	}
	if err := d.Doc.RootMap().Set(key, value); err != nil {
		return err
	}
	_, err := d.Doc.Commit("set")
	if err == nil {
		d.changesSinceCompact++
		d.notifyWatchers()
	}
	return err
}

// Get retrieves a value from the document.
func (d *Document) Get(key string) (interface{}, bool) {
	if d.Doc == nil {
		return nil, false
	}
	v, err := automerge.As[interface{}](d.Doc.RootMap().Get(key))
	if err != nil {
		return nil, false
	}
	return v, true
}

// Map returns the document's contents as a map.
func (d *Document) Map() (map[string]interface{}, error) {
	if d.Doc == nil {
		return nil, nil
	}
	m := make(map[string]interface{})
	keys, err := d.Doc.RootMap().Keys()
	if err != nil {
		return nil, err
	}
	for _, k := range keys {
		v, err := d.Doc.RootMap().Get(k)
		if err != nil {
			return nil, err
		}
		m[k], err = automerge.As[interface{}](v)
		if err != nil {
			return nil, err
		}
	}
	return m, nil
}

// Repo holds a collection of documents, manages storage, and handles peer connections.
type Repo struct {
	ID          RepoID
	docs        map[DocumentID]*Document
	store       StorageAdapter
	sharePolicy SharePolicy
}

// New returns a new empty repository with a random identifier.
func New() *Repo {
	return &Repo{
		ID:          uuid.New(),
		docs:        make(map[DocumentID]*Document),
		sharePolicy: PermissiveSharePolicy{},
	}
}

// NewWithStore creates a repository that will persist documents using the provided store.
func NewWithStore(store StorageAdapter) *Repo {
	r := New()
	r.store = store
	return r
}

// NewDoc creates a new document within the repository and returns it.
func (r *Repo) NewDoc() *Document {
	doc := &Document{ID: uuid.New(), Doc: automerge.New()}
	r.docs[doc.ID] = doc
	return doc
}

// GetDoc retrieves a document by id.
func (r *Repo) GetDoc(id DocumentID) (*Document, bool) {
	d, ok := r.docs[id]
	return d, ok
}

// CompactDoc writes a full snapshot of the document to disk.
func (r *Repo) CompactDoc(id DocumentID) error {
	if r.store == nil {
		return fmt.Errorf("no store configured")
	}
	doc, ok := r.docs[id]
	if !ok {
		return fmt.Errorf("document %s not found", id)
	}
	doc.changesSinceCompact = 0
	return r.store.Compact(doc)
}

// SaveDoc writes a document to disk using the repo's store.
// It will append changes incrementally and compact the document every 10 changes.
func (r *Repo) SaveDoc(id DocumentID) error {
	if r.store == nil {
		return fmt.Errorf("no store configured")
	}
	doc, ok := r.docs[id]
	if !ok {
		return fmt.Errorf("document %s not found", id)
	}
	if doc.changesSinceCompact >= 10 {
		return r.CompactDoc(id)
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

// WithSharePolicy returns a copy of the repo configured to use sp for
// decisions about sharing documents with peers.
func (r *Repo) WithSharePolicy(sp SharePolicy) *Repo {
	r.sharePolicy = sp
	return r
}

// ClearDocs removes all documents from the repo. This is useful for testing.
func (r *Repo) ClearDocs() {
	r.docs = make(map[DocumentID]*Document)
}
