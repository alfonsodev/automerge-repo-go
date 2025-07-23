package repo

import (
	"bytes"
	"testing"

	"github.com/google/uuid"
)

func TestRepoMessageEncodeDecode(t *testing.T) {
	msg := RepoMessage{
		Type:       "sync",
		FromRepoID: New().ID, // just use random id
		ToRepoID:   New().ID,
		DocumentID: uuid.New(),
		Message:    []byte("hello"),
	}
	data, err := msg.Encode()
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	round, err := DecodeRepoMessage(data)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if round.Type != msg.Type || round.FromRepoID != msg.FromRepoID || round.ToRepoID != msg.ToRepoID || round.DocumentID != msg.DocumentID || !bytes.Equal(round.Message, msg.Message) {
		t.Fatalf("round trip mismatch: %+v vs %+v", msg, round)
	}
}
