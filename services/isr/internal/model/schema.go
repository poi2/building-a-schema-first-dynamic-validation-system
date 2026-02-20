package model

import "time"

// Schema represents a schema stored in the registry
type Schema struct {
	ID           string    `db:"id"`
	Version      string    `db:"version"`
	SchemaBinary []byte    `db:"schema_binary"`
	SizeBytes    int32     `db:"size_bytes"`
	CreatedAt    time.Time `db:"created_at"`
}
