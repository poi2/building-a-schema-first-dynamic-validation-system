package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/poi2/building-a-schema-first-dynamic-validation-system/services/isr/internal/model"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	// Read from environment variable with fallback to default
	dbURL := os.Getenv("CELO_DB_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5433/isr?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to test database: %v", err)
	}

	// Drop and recreate table with new schema
	migration := `
		DROP TABLE IF EXISTS schemas;
		CREATE TABLE schemas (
			id VARCHAR(36) PRIMARY KEY,
			version VARCHAR(20) UNIQUE NOT NULL,
			major INTEGER NOT NULL,
			minor INTEGER NOT NULL,
			patch INTEGER NOT NULL,
			schema_binary BYTEA NOT NULL,
			size_bytes INTEGER NOT NULL,
			created_at TIMESTAMP NOT NULL
		);
		CREATE INDEX idx_schemas_semver ON schemas(major DESC, minor DESC, patch DESC);
	`
	_, err = pool.Exec(context.Background(), migration)
	if err != nil {
		t.Skipf("Skipping test: cannot setup database: %v", err)
	}

	return pool
}

func TestSchemaRepository_Create(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	repo := NewSchemaRepository(pool)

	schema := &model.Schema{
		ID:           "test-id-123",
		Version:      "1.0.0",
		Major:        1,
		Minor:        0,
		Patch:        0,
		SchemaBinary: []byte("test binary data"),
		SizeBytes:    16,
		CreatedAt:    time.Now(),
	}

	err := repo.Create(context.Background(), schema)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify it was created
	retrieved, err := repo.GetByVersion(context.Background(), "1.0.0")
	if err != nil {
		t.Fatalf("GetByVersion failed: %v", err)
	}

	if retrieved.ID != schema.ID {
		t.Errorf("Expected ID %s, got %s", schema.ID, retrieved.ID)
	}
	if retrieved.Version != schema.Version {
		t.Errorf("Expected Version %s, got %s", schema.Version, retrieved.Version)
	}
}

func TestSchemaRepository_GetLatestPatch(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	repo := NewSchemaRepository(pool)
	ctx := context.Background()

	// Insert multiple versions
	testVersions := []struct {
		version             string
		major, minor, patch int32
	}{
		{"1.0.0", 1, 0, 0},
		{"1.0.1", 1, 0, 1},
		{"1.0.2", 1, 0, 2},
		{"1.1.0", 1, 1, 0},
	}
	for _, tv := range testVersions {
		schema := &model.Schema{
			ID:           "id-" + tv.version,
			Version:      tv.version,
			Major:        tv.major,
			Minor:        tv.minor,
			Patch:        tv.patch,
			SchemaBinary: []byte("binary-" + tv.version),
			SizeBytes:    int32(len("binary-" + tv.version)),
			CreatedAt:    time.Now(),
		}
		err := repo.Create(ctx, schema)
		if err != nil {
			t.Fatalf("Failed to create schema %s: %v", tv.version, err)
		}
	}

	// Get latest patch for 1.0.x
	latest, err := repo.GetLatestPatch(ctx, 1, 0)
	if err != nil {
		t.Fatalf("GetLatestPatch failed: %v", err)
	}

	if latest.Version != "1.0.2" {
		t.Errorf("Expected version 1.0.2, got %s", latest.Version)
	}
}

func TestSchemaRepository_VersionExists(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	repo := NewSchemaRepository(pool)
	ctx := context.Background()

	// Check non-existent version
	exists, err := repo.VersionExists(ctx, "9.9.9")
	if err != nil {
		t.Fatalf("VersionExists failed: %v", err)
	}
	if exists {
		t.Error("Expected version 9.9.9 to not exist")
	}

	// Create a version
	schema := &model.Schema{
		ID:           "test-id",
		Version:      "2.0.0",
		Major:        2,
		Minor:        0,
		Patch:        0,
		SchemaBinary: []byte("test"),
		SizeBytes:    4,
		CreatedAt:    time.Now(),
	}
	err = repo.Create(ctx, schema)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Check it exists
	exists, err = repo.VersionExists(ctx, "2.0.0")
	if err != nil {
		t.Fatalf("VersionExists failed: %v", err)
	}
	if !exists {
		t.Error("Expected version 2.0.0 to exist")
	}
}
