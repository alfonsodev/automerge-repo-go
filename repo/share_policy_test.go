package repo

import (
	"testing"
	"time"
)

type denyPolicy struct{}

func (denyPolicy) ShouldSync(DocumentID, RepoID) ShareDecision     { return DontShare }
func (denyPolicy) ShouldRequest(DocumentID, RepoID) ShareDecision  { return DontShare }
func (denyPolicy) ShouldAnnounce(DocumentID, RepoID) ShareDecision { return DontShare }

type requestDenyPolicy struct{}

func (requestDenyPolicy) ShouldSync(DocumentID, RepoID) ShareDecision     { return Share }
func (requestDenyPolicy) ShouldRequest(DocumentID, RepoID) ShareDecision  { return DontShare }
func (requestDenyPolicy) ShouldAnnounce(DocumentID, RepoID) ShareDecision { return Share }

type announceDenyPolicy struct{}

func (announceDenyPolicy) ShouldSync(DocumentID, RepoID) ShareDecision     { return Share }
func (announceDenyPolicy) ShouldRequest(DocumentID, RepoID) ShareDecision  { return Share }
func (announceDenyPolicy) ShouldAnnounce(DocumentID, RepoID) ShareDecision { return DontShare }

// Test that sync messages are skipped when the share policy returns DontShare.
func TestSharePolicyBlocksSync(t *testing.T) {
	r1 := New().WithSharePolicy(denyPolicy{})
	r2 := New()
	h1 := NewRepoHandle(r1)
	h2 := NewRepoHandle(r2)

	c1, c2 := newMockConn()
	_ = h1.AddConn(h2.Repo.ID, c1)
	_ = h2.AddConn(h1.Repo.ID, c2)

	doc := h1.Repo.NewDoc()
	if err := doc.Set("k", "v"); err != nil {
		t.Fatalf("set err: %v", err)
	}
	if err := h1.SyncDocument(h2.Repo.ID, doc.ID); err != nil {
		t.Fatalf("sync err: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	if _, ok := h2.Repo.GetDoc(doc.ID); ok {
		t.Fatalf("document should not be shared")
	}

	h1.Close()
	h2.Close()
}

// Test that a document is not created when ShouldRequest returns DontShare.
func TestSharePolicyBlocksRequest(t *testing.T) {
	r1 := New()
	r2 := New().WithSharePolicy(requestDenyPolicy{})
	h1 := NewRepoHandle(r1)
	h2 := NewRepoHandle(r2)

	c1, c2 := newMockConn()
	_ = h1.AddConn(h2.Repo.ID, c1)
	_ = h2.AddConn(h1.Repo.ID, c2)

	doc := h1.Repo.NewDoc()
	if err := doc.Set("k", "v"); err != nil {
		t.Fatalf("set err: %v", err)
	}
	if err := h1.SyncDocument(h2.Repo.ID, doc.ID); err != nil {
		t.Fatalf("sync err: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	if _, ok := h2.Repo.GetDoc(doc.ID); ok {
		t.Fatalf("document should not have been requested")
	}

	h1.Close()
	h2.Close()
}

// Test that SyncAll checks ShouldAnnounce before sending documents.
func TestSharePolicyBlocksAnnounce(t *testing.T) {
	r1 := New().WithSharePolicy(announceDenyPolicy{})
	r2 := New()
	h1 := NewRepoHandle(r1)
	h2 := NewRepoHandle(r2)

	c1, c2 := newMockConn()
	_ = h1.AddConn(h2.Repo.ID, c1)
	_ = h2.AddConn(h1.Repo.ID, c2)

	doc := h1.Repo.NewDoc()
	if err := doc.Set("k", "v"); err != nil {
		t.Fatalf("set err: %v", err)
	}
	if err := h1.SyncAll(h2.Repo.ID); err != nil {
		t.Fatalf("sync err: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	if _, ok := h2.Repo.GetDoc(doc.ID); ok {
		t.Fatalf("document should not have been announced")
	}

	h1.Close()
	h2.Close()
}
