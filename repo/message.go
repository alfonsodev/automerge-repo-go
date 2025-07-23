package repo

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"
	"github.com/google/uuid"
)

// RepoMessage represents a sync or ephemeral message exchanged between repositories.
// Type should be either "sync" or "ephemeral".
type RepoMessage struct {
	Type       string // "sync" or "ephemeral"
	FromRepoID RepoID
	ToRepoID   RepoID
	DocumentID DocumentID
	Message    []byte
}

// repoMessageCBOR mirrors the on-the-wire CBOR structure.
// IDs are encoded as strings for compatibility with other implementations.
type repoMessageCBOR struct {
	Type       string `cbor:"type"`
	SenderID   string `cbor:"senderId"`
	TargetID   string `cbor:"targetId"`
	DocumentID string `cbor:"documentId"`
	Message    []byte `cbor:"message"`
}

// Encode converts the RepoMessage into CBOR bytes for transmission.
func (m RepoMessage) Encode() ([]byte, error) {
	if m.Type != "sync" && m.Type != "ephemeral" {
		return nil, fmt.Errorf("invalid RepoMessage type %q", m.Type)
	}
	wire := repoMessageCBOR{
		Type:       m.Type,
		SenderID:   m.FromRepoID.String(),
		TargetID:   m.ToRepoID.String(),
		DocumentID: m.DocumentID.String(),
		Message:    m.Message,
	}
	return cbor.Marshal(wire)
}

// DecodeRepoMessage parses CBOR data into a RepoMessage.
func DecodeRepoMessage(data []byte) (RepoMessage, error) {
	var wire repoMessageCBOR
	if err := cbor.Unmarshal(data, &wire); err != nil {
		return RepoMessage{}, err
	}
	if wire.Type != "sync" && wire.Type != "ephemeral" {
		return RepoMessage{}, fmt.Errorf("invalid RepoMessage type %q", wire.Type)
	}
	from, err := uuid.Parse(wire.SenderID)
	if err != nil {
		return RepoMessage{}, err
	}
	to, err := uuid.Parse(wire.TargetID)
	if err != nil {
		return RepoMessage{}, err
	}
	doc, err := uuid.Parse(wire.DocumentID)
	if err != nil {
		return RepoMessage{}, err
	}
	return RepoMessage{
		Type:       wire.Type,
		FromRepoID: from,
		ToRepoID:   to,
		DocumentID: doc,
		Message:    wire.Message,
	}, nil
}
