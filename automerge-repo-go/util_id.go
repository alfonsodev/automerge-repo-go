//go:build !js

package repo

import "github.com/google/uuid"

// parseRepoID attempts to parse the given string as a UUID. If parsing fails
// (for example when the sender uses non-UUID identifiers like "peer-rgp224jx")
// it deterministically generates a UUID v5 (SHA-1) from the original string.
//
// This allows interop with JS implementations that use human-readable peer IDs
// while keeping the rest of the codebase working with the uuid.UUID type.
func parseRepoID(s string) uuid.UUID {
    if id, err := uuid.Parse(s); err == nil {
        return id
    }
    // Use the string to derive a deterministic UUID so that the same peer ID
    // maps to the same UUID across connections.
    return uuid.NewSHA1(uuid.NameSpaceOID, []byte(s))
} 