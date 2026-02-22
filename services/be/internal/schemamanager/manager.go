package schemamanager

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"connectrpc.com/connect"
	isrv1 "github.com/poi2/building-a-schema-first-dynamic-validation-system/pkg/gen/go/isr/v1"
	"github.com/poi2/building-a-schema-first-dynamic-validation-system/pkg/gen/go/isr/v1/isrv1connect"
	"github.com/poi2/building-a-schema-first-dynamic-validation-system/services/be/internal/validator"
)

// SchemaManager manages schema updates from ISR
type SchemaManager struct {
	config    Config
	validator *validator.SchemaAwareValidator
	client    isrv1connect.SchemaRegistryServiceClient
	stopCh    chan struct{}
	doneCh    chan struct{}
}

// NewSchemaManager creates a new schema manager
func NewSchemaManager(config Config, validator *validator.SchemaAwareValidator) *SchemaManager {
	client := isrv1connect.NewSchemaRegistryServiceClient(
		http.DefaultClient,
		config.ISRURL,
	)

	return &SchemaManager{
		config:    config,
		validator: validator,
		client:    client,
		stopCh:    make(chan struct{}),
		doneCh:    make(chan struct{}),
	}
}

// LoadInitialSchema loads the initial schema from ISR
func (m *SchemaManager) LoadInitialSchema(ctx context.Context) error {
	req := connect.NewRequest(&isrv1.GetLatestPatchRequest{
		Major: m.config.Major,
		Minor: m.config.Minor,
	})

	resp, err := m.client.GetLatestPatch(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to get initial schema from ISR: %w", err)
	}

	version := resp.Msg.Metadata.Version
	if err := m.validator.UpdateSchema(resp.Msg.SchemaBinary, version); err != nil {
		return fmt.Errorf("failed to initialize validator with schema: %w", err)
	}

	log.Printf("Schema initialized: target=%s, loaded version=%s", m.config.SchemaTarget, version)
	return nil
}

// Start starts the schema polling goroutine
func (m *SchemaManager) Start(ctx context.Context) {
	go m.pollLoop(ctx)
}

// Stop gracefully stops the schema manager
func (m *SchemaManager) Stop() {
	close(m.stopCh)
	<-m.doneCh
	log.Println("Schema manager stopped")
}

// pollLoop runs the schema polling loop
func (m *SchemaManager) pollLoop(ctx context.Context) {
	defer close(m.doneCh)

	ticker := time.NewTicker(m.config.PollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := m.checkAndUpdateSchema(ctx); err != nil {
				log.Printf("Schema polling error (will retry in %s): %v",
					m.config.PollingInterval, err)
			}
		case <-m.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

// checkAndUpdateSchema checks for schema updates and performs hot-swap if needed
func (m *SchemaManager) checkAndUpdateSchema(ctx context.Context) error {
	currentVersion := m.validator.GetCurrentVersion()

	req := connect.NewRequest(&isrv1.GetLatestPatchRequest{
		Major: m.config.Major,
		Minor: m.config.Minor,
	})

	resp, err := m.client.GetLatestPatch(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to get latest patch from ISR: %w", err)
	}

	latestVersion := resp.Msg.Metadata.Version

	if currentVersion == latestVersion {
		return nil // No update needed
	}

	if err := m.validator.UpdateSchema(resp.Msg.SchemaBinary, latestVersion); err != nil {
		return fmt.Errorf("failed to update schema: %w", err)
	}

	log.Printf("Hot-swapped validator: %s -> %s", currentVersion, latestVersion)
	return nil
}
