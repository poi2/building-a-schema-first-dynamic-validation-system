package model

import "time"

// Schema represents a schema stored in the registry
type Schema struct {
	ID           string    `db:"id"`
	Version      string    `db:"version"`
	Major        int32     `db:"major"`
	Minor        int32     `db:"minor"`
	Patch        int32     `db:"patch"`
	SchemaBinary []byte    `db:"schema_binary"`
	SizeBytes    int32     `db:"size_bytes"`
	CreatedAt    time.Time `db:"created_at"`
}
