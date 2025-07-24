package repo

import (
	automerge "github.com/automerge/automerge-go"
)

// DocumentHandle provides access to a document along with a mechanism
// to wait for changes.
type DocumentHandle struct {
	doc *Document
}

// Changed returns a channel that will receive a single notification when
// the document is next modified.
func (h *DocumentHandle) Changed() <-chan struct{} {
	return h.doc.watch()
}

// WithDoc runs f with the underlying Automerge document.
func (h *DocumentHandle) WithDoc(f func(*automerge.Doc)) {
	h.doc.ensureDoc()
	f(h.doc.doc)
}

// WithDocMut runs f with the document and commits the result. A change
// notification is sent if the document was modified.
func (h *DocumentHandle) WithDocMut(f func(*automerge.Doc) error) error {
	h.doc.ensureDoc()
	if err := f(h.doc.doc); err != nil {
		return err
	}
	_, err := h.doc.doc.Commit("update")
	if err == nil {
		h.doc.notifyWatchers()
	}
	return err
}

// NewDocHandle creates a new document and returns a handle to it.
func (r *Repo) NewDocHandle() *DocumentHandle {
	doc := r.NewDoc()
	return &DocumentHandle{doc: doc}
}

// GetDocHandle returns a handle for the document with the given id.
func (r *Repo) GetDocHandle(id DocumentID) (*DocumentHandle, bool) {
	d, ok := r.GetDoc(id)
	if !ok {
		return nil, false
	}
	return &DocumentHandle{doc: d}, true
}

// --- internal helpers on Document ---

func (d *Document) ensureDoc() {
	if d.doc == nil {
		d.doc = automerge.New()
	}
}

func (d *Document) watch() <-chan struct{} {
	ch := make(chan struct{}, 1)
	d.watchersMu.Lock()
	d.watchers = append(d.watchers, ch)
	d.watchersMu.Unlock()
	return ch
}

func (d *Document) notifyWatchers() {
	d.watchersMu.Lock()
	w := d.watchers
	d.watchers = nil
	d.watchersMu.Unlock()
	for _, ch := range w {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}
