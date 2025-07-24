package repo

// ShareDecision indicates whether to share or withhold a document.
type ShareDecision int

const (
	Share ShareDecision = iota
	DontShare
)

// SharePolicy decides how documents should be shared with peers.
// All methods should return Share or DontShare to indicate the policy
// decision.
type SharePolicy interface {
	ShouldSync(docID DocumentID, withPeer RepoID) ShareDecision
	ShouldRequest(docID DocumentID, fromPeer RepoID) ShareDecision
	ShouldAnnounce(docID DocumentID, toPeer RepoID) ShareDecision
}

// PermissiveSharePolicy shares all documents with all peers.
type PermissiveSharePolicy struct{}

func (PermissiveSharePolicy) ShouldSync(DocumentID, RepoID) ShareDecision     { return Share }
func (PermissiveSharePolicy) ShouldRequest(DocumentID, RepoID) ShareDecision  { return Share }
func (PermissiveSharePolicy) ShouldAnnounce(DocumentID, RepoID) ShareDecision { return Share }
