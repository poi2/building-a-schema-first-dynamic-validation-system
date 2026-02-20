package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/poi2/building-a-schema-first-dynamic-validation-system/services/isr/internal/model"
)

type SchemaRepository struct {
	pool *pgxpool.Pool
}

func NewSchemaRepository(pool *pgxpool.Pool) *SchemaRepository {
	return &SchemaRepository{pool: pool}
}

// Create inserts a new schema into the database
func (r *SchemaRepository) Create(ctx context.Context, schema *model.Schema) error {
	query := `
		INSERT INTO schemas (id, version, major, minor, patch, schema_binary, size_bytes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.pool.Exec(ctx, query,
		schema.ID,
		schema.Version,
		schema.Major,
		schema.Minor,
		schema.Patch,
		schema.SchemaBinary,
		schema.SizeBytes,
		schema.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert schema: %w", err)
	}
	return nil
}

// GetByVersion retrieves a schema by its version
func (r *SchemaRepository) GetByVersion(ctx context.Context, version string) (*model.Schema, error) {
	query := `
		SELECT id, version, major, minor, patch, schema_binary, size_bytes, created_at
		FROM schemas
		WHERE version = $1
	`
	var schema model.Schema
	err := r.pool.QueryRow(ctx, query, version).Scan(
		&schema.ID,
		&schema.Version,
		&schema.Major,
		&schema.Minor,
		&schema.Patch,
		&schema.SchemaBinary,
		&schema.SizeBytes,
		&schema.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema by version: %w", err)
	}
	return &schema, nil
}

// GetLatestPatch retrieves the latest patch version for a given major.minor
func (r *SchemaRepository) GetLatestPatch(ctx context.Context, major, minor int32) (*model.Schema, error) {
	query := `
		SELECT id, version, major, minor, patch, schema_binary, size_bytes, created_at
		FROM schemas
		WHERE major = $1 AND minor = $2
		ORDER BY patch DESC
		LIMIT 1
	`

	var schema model.Schema
	err := r.pool.QueryRow(ctx, query, major, minor).Scan(
		&schema.ID,
		&schema.Version,
		&schema.Major,
		&schema.Minor,
		&schema.Patch,
		&schema.SchemaBinary,
		&schema.SizeBytes,
		&schema.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest patch: %w", err)
	}
	return &schema, nil
}

// VersionExists checks if a version already exists
func (r *SchemaRepository) VersionExists(ctx context.Context, version string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM schemas WHERE version = $1)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, version).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check version existence: %w", err)
	}
	return exists, nil
}
