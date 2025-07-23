package repo

import (
	"fmt"

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
