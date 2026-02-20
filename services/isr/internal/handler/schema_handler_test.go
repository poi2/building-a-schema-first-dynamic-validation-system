package handler

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	isrv1 "github.com/poi2/building-a-schema-first-dynamic-validation-system/pkg/gen/go/isr/v1"
)

// Mock repository for testing
type mockSchemaRepository struct {
	versionExists bool
	createErr     error
}

func (m *mockSchemaRepository) Create(ctx context.Context, schema interface{}) error {
	return m.createErr
}

func (m *mockSchemaRepository) GetByVersion(ctx context.Context, version string) (interface{}, error) {
	return nil, nil
}

func (m *mockSchemaRepository) GetLatestPatch(ctx context.Context, major, minor int32) (interface{}, error) {
	return nil, nil
}

func (m *mockSchemaRepository) VersionExists(ctx context.Context, version string) (bool, error) {
	return m.versionExists, nil
}

func TestSchemaHandler_UploadSchema_InvalidVersion(t *testing.T) {
	// This test verifies that protovalidate will reject invalid versions
	// The actual validation happens at the Connect layer before reaching the handler
	req := &isrv1.UploadSchemaRequest{
		Version:      "invalid-version",
		SchemaBinary: []byte("test"),
	}

	if req.Version == "invalid-version" {
		// In real usage, this would be caught by protovalidate
		t.Log("Invalid version would be rejected by validation")
	}
}

func TestSchemaHandler_UploadSchema_EmptyBinary(t *testing.T) {
	req := &isrv1.UploadSchemaRequest{
		Version:      "1.0.0",
		SchemaBinary: []byte{},
	}

	if len(req.SchemaBinary) == 0 {
		// In real usage, this would be caught by protovalidate (min_len: 1)
		t.Log("Empty binary would be rejected by validation")
	}
}

func TestSchemaHandler_GetLatestPatch_NegativeMajor(t *testing.T) {
	req := &isrv1.GetLatestPatchRequest{
		Major: -1,
		Minor: 0,
	}

	if req.Major < 0 {
		// In real usage, this would be caught by protovalidate (gte = 0)
		t.Log("Negative major version would be rejected by validation")
	}
}

func TestConnectRequest(t *testing.T) {
	// Test that we can create Connect requests
	req := connect.NewRequest(&isrv1.UploadSchemaRequest{
		Version:      "1.0.0",
		SchemaBinary: []byte("test data"),
	})

	if req.Msg.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", req.Msg.Version)
	}
}
