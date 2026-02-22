package schemamanager

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
	isrv1 "github.com/poi2/building-a-schema-first-dynamic-validation-system/pkg/gen/go/isr/v1"
	"github.com/poi2/building-a-schema-first-dynamic-validation-system/pkg/gen/go/isr/v1/isrv1connect"
	"github.com/poi2/building-a-schema-first-dynamic-validation-system/services/be/internal/validator"
	"google.golang.org/protobuf/proto"

	// Import generated proto to register file descriptors
	_ "github.com/poi2/building-a-schema-first-dynamic-validation-system/pkg/gen/go/user/v1"
)

// mockISRServer creates a mock ISR server for testing
type mockISRServer struct {
	isrv1connect.UnimplementedSchemaRegistryServiceHandler
	version        string
	descriptorData []byte
	errorToReturn  error
}

func (m *mockISRServer) GetLatestPatch(
	ctx context.Context,
	req *connect.Request[isrv1.GetLatestPatchRequest],
) (*connect.Response[isrv1.GetLatestPatchResponse], error) {
	if m.errorToReturn != nil {
		return nil, m.errorToReturn
	}

	return connect.NewResponse(&isrv1.GetLatestPatchResponse{
		Metadata: &isrv1.SchemaMetadata{
			Version: m.version,
		},
		SchemaBinary: m.descriptorData,
	}), nil
}

func setupMockISRServer(t *testing.T, version string, shouldError bool) (*httptest.Server, []byte) {
	t.Helper()

	descriptorData := validator.CreateTestDescriptorBytes(t)

	mock := &mockISRServer{
		version:        version,
		descriptorData: descriptorData,
	}

	if shouldError {
		mock.errorToReturn = errors.New("ISR service error")
	}

	mux := http.NewServeMux()
	path, handler := isrv1connect.NewSchemaRegistryServiceHandler(mock)
	mux.Handle(path, handler)

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	return server, descriptorData
}

func TestSchemaManager_LoadInitialSchema_Success(t *testing.T) {
	server, _ := setupMockISRServer(t, "1.0.0", false)

	config := Config{
		ISRURL:          server.URL,
		SchemaTarget:    "1.0",
		Major:           1,
		Minor:           0,
		PollingInterval: 1 * time.Minute,
	}

	schemaValidator := &validator.SchemaAwareValidator{}
	manager := NewSchemaManager(config, schemaValidator)

	ctx := context.Background()
	err := manager.LoadInitialSchema(ctx)
	if err != nil {
		t.Fatalf("LoadInitialSchema failed: %v", err)
	}

	version := schemaValidator.GetCurrentVersion()
	if version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", version)
	}
}

func TestSchemaManager_LoadInitialSchema_ISRError(t *testing.T) {
	server, _ := setupMockISRServer(t, "1.0.0", true)

	config := Config{
		ISRURL:          server.URL,
		SchemaTarget:    "1.0",
		Major:           1,
		Minor:           0,
		PollingInterval: 1 * time.Minute,
	}

	schemaValidator := &validator.SchemaAwareValidator{}
	manager := NewSchemaManager(config, schemaValidator)

	ctx := context.Background()
	err := manager.LoadInitialSchema(ctx)
	if err == nil {
		t.Fatal("expected error from ISR, got nil")
	}
}

func TestSchemaManager_CheckAndUpdateSchema_NoUpdate(t *testing.T) {
	server, _ := setupMockISRServer(t, "1.0.0", false)

	config := Config{
		ISRURL:          server.URL,
		SchemaTarget:    "1.0",
		Major:           1,
		Minor:           0,
		PollingInterval: 1 * time.Minute,
	}

	schemaValidator := &validator.SchemaAwareValidator{}
	manager := NewSchemaManager(config, schemaValidator)

	ctx := context.Background()

	// Load initial schema
	if err := manager.LoadInitialSchema(ctx); err != nil {
		t.Fatalf("LoadInitialSchema failed: %v", err)
	}

	initialVersion := schemaValidator.GetCurrentVersion()

	// Check for updates (should find none)
	if err := manager.checkAndUpdateSchema(ctx); err != nil {
		t.Fatalf("checkAndUpdateSchema failed: %v", err)
	}

	newVersion := schemaValidator.GetCurrentVersion()
	if newVersion != initialVersion {
		t.Errorf("version should not have changed: %s -> %s", initialVersion, newVersion)
	}
}

func TestSchemaManager_CheckAndUpdateSchema_HotSwap(t *testing.T) {
	// Start with version 1.0.0
	server, descriptorData := setupMockISRServer(t, "1.0.0", false)

	config := Config{
		ISRURL:          server.URL,
		SchemaTarget:    "1.0",
		Major:           1,
		Minor:           0,
		PollingInterval: 1 * time.Minute,
	}

	schemaValidator := &validator.SchemaAwareValidator{}
	manager := NewSchemaManager(config, schemaValidator)

	ctx := context.Background()

	// Load initial schema
	if err := manager.LoadInitialSchema(ctx); err != nil {
		t.Fatalf("LoadInitialSchema failed: %v", err)
	}

	initialVersion := schemaValidator.GetCurrentVersion()
	if initialVersion != "1.0.0" {
		t.Errorf("expected initial version 1.0.0, got %s", initialVersion)
	}

	// Update mock server to return new version
	server.Config.Handler.(*http.ServeMux).HandleFunc(isrv1connect.SchemaRegistryServiceGetLatestPatchProcedure, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/proto")
		resp := &isrv1.GetLatestPatchResponse{
			Metadata: &isrv1.SchemaMetadata{
				Version: "1.0.1",
			},
			SchemaBinary: descriptorData,
		}
		data, _ := proto.Marshal(resp)
		w.Write(data)
	})

	// Recreate client and manager to pick up new handler
	manager = NewSchemaManager(config, schemaValidator)

	// Check for updates (should find new version)
	if err := manager.checkAndUpdateSchema(ctx); err != nil {
		t.Fatalf("checkAndUpdateSchema failed: %v", err)
	}

	newVersion := schemaValidator.GetCurrentVersion()
	if newVersion != "1.0.1" {
		t.Errorf("expected new version 1.0.1, got %s", newVersion)
	}
}

func TestSchemaManager_CheckAndUpdateSchema_InvalidSchema(t *testing.T) {
	server, _ := setupMockISRServer(t, "1.0.0", false)

	config := Config{
		ISRURL:          server.URL,
		SchemaTarget:    "1.0",
		Major:           1,
		Minor:           0,
		PollingInterval: 1 * time.Minute,
	}

	schemaValidator := &validator.SchemaAwareValidator{}
	manager := NewSchemaManager(config, schemaValidator)

	ctx := context.Background()

	// Load initial schema
	if err := manager.LoadInitialSchema(ctx); err != nil {
		t.Fatalf("LoadInitialSchema failed: %v", err)
	}

	initialVersion := schemaValidator.GetCurrentVersion()

	// Update mock server to return invalid schema
	server.Config.Handler.(*http.ServeMux).HandleFunc(isrv1connect.SchemaRegistryServiceGetLatestPatchProcedure, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/proto")
		resp := &isrv1.GetLatestPatchResponse{
			Metadata: &isrv1.SchemaMetadata{
				Version: "1.0.1",
			},
			SchemaBinary: []byte("invalid schema data"),
		}
		data, _ := proto.Marshal(resp)
		w.Write(data)
	})

	// Recreate client and manager to pick up new handler
	manager = NewSchemaManager(config, schemaValidator)

	// Check for updates (should fail due to invalid schema)
	err := manager.checkAndUpdateSchema(ctx)
	if err == nil {
		t.Fatal("expected error for invalid schema, got nil")
	}

	// Version should not have changed
	currentVersion := schemaValidator.GetCurrentVersion()
	if currentVersion != initialVersion {
		t.Errorf("version should not have changed on error: %s -> %s", initialVersion, currentVersion)
	}
}

func TestSchemaManager_StartStop(t *testing.T) {
	server, _ := setupMockISRServer(t, "1.0.0", false)

	config := Config{
		ISRURL:          server.URL,
		SchemaTarget:    "1.0",
		Major:           1,
		Minor:           0,
		PollingInterval: 100 * time.Millisecond,
	}

	schemaValidator := &validator.SchemaAwareValidator{}
	manager := NewSchemaManager(config, schemaValidator)

	ctx := context.Background()

	// Load initial schema
	if err := manager.LoadInitialSchema(ctx); err != nil {
		t.Fatalf("LoadInitialSchema failed: %v", err)
	}

	// Start manager
	manager.Start(ctx)

	// Wait a bit to ensure polling has started
	time.Sleep(150 * time.Millisecond)

	// Stop manager
	manager.Stop()

	// Verify manager stopped gracefully
	// (doneCh should be closed, so this should not block)
	select {
	case <-manager.doneCh:
		// Success
	case <-time.After(1 * time.Second):
		t.Error("manager did not stop within timeout")
	}
}
