package repo

import (
	"testing"
	"time"
)

type denyPolicy struct{}

func (denyPolicy) ShouldSync(DocumentID, RepoID) ShareDecision     { return DontShare }
func (denyPolicy) ShouldRequest(DocumentID, RepoID) ShareDecision  { return DontShare }
func (denyPolicy) ShouldAnnounce(DocumentID, RepoID) ShareDecision { return DontShare }

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
