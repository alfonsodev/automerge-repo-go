package repo

import "testing"

// denyPolicy rejects all share actions
type denyPolicy struct{}

func (denyPolicy) ShouldSync(DocumentID, RepoID) ShareDecision     { return DontShare }
func (denyPolicy) ShouldRequest(DocumentID, RepoID) ShareDecision  { return DontShare }
func (denyPolicy) ShouldAnnounce(DocumentID, RepoID) ShareDecision { return DontShare }

func TestSharePolicyBlocksSync(t *testing.T) {
	r1 := New()
	r2 := New()
	r1.SetSharePolicy(denyPolicy{})

	h1 := NewRepoHandle(r1)
	h2 := NewRepoHandle(r2)

	c1, c2 := newMockConn()
	_ = h1.AddConn(h2.Repo.ID, c1)
	_ = h2.AddConn(h1.Repo.ID, c2)

	doc := h1.Repo.NewDoc()
	_ = doc.Set("k", "v")

	if err := h1.SyncDocument(h2.Repo.ID, doc.ID); err != nil {
		t.Fatalf("sync err: %v", err)
	}

	select {
	case <-c1.sendCh:
		t.Fatal("unexpected message sent")
	default:
	}

	h1.Close()
	h2.Close()
}
