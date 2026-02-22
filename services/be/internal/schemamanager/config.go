package schemamanager

import "time"

// Config holds the configuration for the schema manager
type Config struct {
	// ISRURL is the URL of the ISR service (e.g., "http://localhost:50051")
	ISRURL string

	// SchemaTarget is the target schema version in "Major.Minor" format (e.g., "1.0")
	SchemaTarget string

	// Major is the major version number
	Major int32

	// Minor is the minor version number
	Minor int32

	// PollingInterval is the interval between schema update checks
	PollingInterval time.Duration
}
