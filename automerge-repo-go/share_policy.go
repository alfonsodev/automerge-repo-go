package repo

// ShareDecision indicates whether to share document updates with a peer.
type ShareDecision int

const (
	// Share indicates that the document should be shared.
	Share ShareDecision = iota
	// DontShare indicates that the document should not be shared.
	DontShare
)

// SharePolicy determines if documents should be shared or requested from peers.
// Implementations may inspect the document and peer IDs to make a decision.
type SharePolicy interface {
	// ShouldSync is consulted before sending or applying a sync message.
	ShouldSync(docID DocumentID, peer RepoID) ShareDecision
	// ShouldRequest decides if the document should be requested from the peer.
	ShouldRequest(docID DocumentID, peer RepoID) ShareDecision
	// ShouldAnnounce decides if we should announce the document to the peer.
	ShouldAnnounce(docID DocumentID, peer RepoID) ShareDecision
}

// PermissiveSharePolicy always allows sharing.
type PermissiveSharePolicy struct{}

// ShouldSync always returns Share.
func (PermissiveSharePolicy) ShouldSync(DocumentID, RepoID) ShareDecision { return Share }

// ShouldRequest always returns Share.
func (PermissiveSharePolicy) ShouldRequest(DocumentID, RepoID) ShareDecision { return Share }

// ShouldAnnounce always returns Share.
func (PermissiveSharePolicy) ShouldAnnounce(DocumentID, RepoID) ShareDecision { return Share }