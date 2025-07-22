package repo

import (
	"github.com/google/uuid"
)

// DocumentID uniquely identifies a document.
type DocumentID = uuid.UUID

// RepoID uniquely identifies a repository instance.
type RepoID = uuid.UUID

// Document represents a single Automerge document.
type Document struct {
	ID   DocumentID
	Data []byte // placeholder for encoded document
}

// Repo holds a collection of documents.
type Repo struct {
	ID   RepoID
	docs map[DocumentID]*Document
}

// New returns a new empty repository with a random identifier.
func New() *Repo {
	return &Repo{ID: uuid.New(), docs: make(map[DocumentID]*Document)}
}

// NewDoc creates a new document within the repository and returns it.
func (r *Repo) NewDoc() *Document {
	doc := &Document{ID: uuid.New()}
	r.docs[doc.ID] = doc
	return doc
}

// GetDoc retrieves a document by id.
func (r *Repo) GetDoc(id DocumentID) (*Document, bool) {
	d, ok := r.docs[id]
	return d, ok
}
