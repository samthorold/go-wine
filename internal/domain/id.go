package domain

import (
	"crypto/rand"
	"encoding/hex"
)

// ID identifies any entity. It is generated in the domain so that the in-memory
// and Postgres adapters store the same value rather than each minting their own.
type ID string

// NewID returns a fresh random ID.
func NewID() ID {
	b := make([]byte, 16)
	// crypto/rand.Read never returns an error on the platforms we target.
	_, _ = rand.Read(b)
	return ID(hex.EncodeToString(b))
}

func (id ID) String() string { return string(id) }
