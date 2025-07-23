package repo

import (
	"encoding/json"
	"fmt"
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

// repoMessageJSON mirrors the on-the-wire JSON structure.
type repoMessageJSON struct {
	Type       string     `json:"type"`
	SenderID   RepoID     `json:"senderId"`
	TargetID   RepoID     `json:"targetId"`
	DocumentID DocumentID `json:"documentId"`
	Message    []byte     `json:"message"`
}

// Encode converts the RepoMessage into JSON bytes for transmission.
func (m RepoMessage) Encode() ([]byte, error) {
	if m.Type != "sync" && m.Type != "ephemeral" {
		return nil, fmt.Errorf("invalid RepoMessage type %q", m.Type)
	}
	wire := repoMessageJSON{
		Type:       m.Type,
		SenderID:   m.FromRepoID,
		TargetID:   m.ToRepoID,
		DocumentID: m.DocumentID,
		Message:    m.Message,
	}
	return json.Marshal(wire)
}

// DecodeRepoMessage parses JSON data into a RepoMessage.
func DecodeRepoMessage(data []byte) (RepoMessage, error) {
	var wire repoMessageJSON
	if err := json.Unmarshal(data, &wire); err != nil {
		return RepoMessage{}, err
	}
	if wire.Type != "sync" && wire.Type != "ephemeral" {
		return RepoMessage{}, fmt.Errorf("invalid RepoMessage type %q", wire.Type)
	}
	return RepoMessage{
		Type:       wire.Type,
		FromRepoID: wire.SenderID,
		ToRepoID:   wire.TargetID,
		DocumentID: wire.DocumentID,
		Message:    wire.Message,
	}, nil
}
