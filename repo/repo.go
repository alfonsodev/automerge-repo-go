package repo

import (
	"fmt"

	"github.com/google/uuid"
)

// DocumentID uniquely identifies a document.
type DocumentID = uuid.UUID

// RepoID uniquely identifies a repository instance.
type RepoID = uuid.UUID

// Document represents a single Automerge document.
type Document struct {
	ID   DocumentID
	Data map[string]interface{}
}

// Set assigns a value in the document.
func (d *Document) Set(key string, value interface{}) {
	if d.Data == nil {
		d.Data = make(map[string]interface{})
	}
	d.Data[key] = value
}

// Get retrieves a value from the document.
func (d *Document) Get(key string) (interface{}, bool) {
	v, ok := d.Data[key]
	return v, ok
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
	doc := &Document{ID: uuid.New(), Data: make(map[string]interface{})}
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
