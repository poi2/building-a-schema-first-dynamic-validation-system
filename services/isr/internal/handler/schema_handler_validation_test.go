package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"connectrpc.com/validate"
	isrv1 "github.com/poi2/building-a-schema-first-dynamic-validation-system/pkg/gen/go/isr/v1"
	"github.com/poi2/building-a-schema-first-dynamic-validation-system/pkg/gen/go/isr/v1/isrv1connect"
	"github.com/poi2/building-a-schema-first-dynamic-validation-system/services/isr/internal/model"
)

const (
	// maxSchemaBinarySize is the maximum allowed schema binary size (10MB)
	maxSchemaBinarySize = 10 * 1024 * 1024 // 10485760 bytes
)

// newTestClient creates a test Connect client with validation interceptor
func newTestClient(t *testing.T, handler *SchemaHandler) (isrv1connect.SchemaRegistryServiceClient, func()) {
	t.Helper()

	mux := http.NewServeMux()
	interceptors := connect.WithInterceptors(
		validate.NewInterceptor(),
	)
	path, connectHandler := isrv1connect.NewSchemaRegistryServiceHandler(handler, interceptors)
	mux.Handle(path, connectHandler)

	server := httptest.NewServer(mux)
	client := isrv1connect.NewSchemaRegistryServiceClient(
		http.DefaultClient,
		server.URL,
	)

	return client, server.Close
}

// TestUploadSchema_ValidationError_SchemaBinaryTooLarge tests that schema binaries larger than 10MB are rejected
func TestUploadSchema_ValidationError_SchemaBinaryTooLarge(t *testing.T) {
	mockRepo := &mockSchemaRepository{
		versionExistsFunc: func(ctx context.Context, version string) (bool, error) {
			return false, nil
		},
		createFunc: func(ctx context.Context, schema *model.Schema) error {
			return nil
		},
	}

	handler := NewSchemaHandler(mockRepo)
	client, cleanup := newTestClient(t, handler)
	defer cleanup()

	// Create a binary larger than 10MB
	largeBinary := make([]byte, maxSchemaBinarySize+1)

	req := connect.NewRequest(&isrv1.UploadSchemaRequest{
		Version:      "1.0.0",
		SchemaBinary: largeBinary,
	})

	_, err := client.UploadSchema(context.Background(), req)
	if err == nil {
		t.Fatal("UploadSchema() with too large binary should fail, but got nil error")
	}

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("error type = %T, want *connect.Error", err)
	}
	if connectErr.Code() != connect.CodeInvalidArgument {
		t.Errorf("error code = %v, want %v (InvalidArgument)", connectErr.Code(), connect.CodeInvalidArgument)
	}
}

// TestUploadSchema_ValidationError_SchemaBinaryEmpty tests that empty schema binaries are rejected
func TestUploadSchema_ValidationError_SchemaBinaryEmpty(t *testing.T) {
	mockRepo := &mockSchemaRepository{}
	handler := NewSchemaHandler(mockRepo)
	client, cleanup := newTestClient(t, handler)
	defer cleanup()

	req := connect.NewRequest(&isrv1.UploadSchemaRequest{
		Version:      "1.0.0",
		SchemaBinary: []byte{}, // Empty binary
	})

	_, err := client.UploadSchema(context.Background(), req)
	if err == nil {
		t.Fatal("UploadSchema() with empty binary should fail, but got nil error")
	}

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("error type = %T, want *connect.Error", err)
	}
	if connectErr.Code() != connect.CodeInvalidArgument {
		t.Errorf("error code = %v, want %v (InvalidArgument)", connectErr.Code(), connect.CodeInvalidArgument)
	}
}

// TestUploadSchema_ValidationSuccess_MaxSize tests that a schema binary at exactly 10MB is accepted
func TestUploadSchema_ValidationSuccess_MaxSize(t *testing.T) {
	mockRepo := &mockSchemaRepository{
		versionExistsFunc: func(ctx context.Context, version string) (bool, error) {
			return false, nil
		},
		createFunc: func(ctx context.Context, schema *model.Schema) error {
			return nil
		},
	}

	handler := NewSchemaHandler(mockRepo)
	client, cleanup := newTestClient(t, handler)
	defer cleanup()

	// Create a binary exactly at the 10MB limit
	maxBinary := make([]byte, maxSchemaBinarySize)

	req := connect.NewRequest(&isrv1.UploadSchemaRequest{
		Version:      "1.0.0",
		SchemaBinary: maxBinary,
	})

	resp, err := client.UploadSchema(context.Background(), req)
	if err != nil {
		t.Fatalf("UploadSchema() with exactly 10MB binary should succeed, but got error: %v", err)
	}

	if resp.Msg.Metadata.Version != "1.0.0" {
		t.Errorf("Version = %v, want 1.0.0", resp.Msg.Metadata.Version)
	}
	if resp.Msg.Metadata.SizeBytes != int32(len(maxBinary)) {
		t.Errorf("SizeBytes = %v, want %v", resp.Msg.Metadata.SizeBytes, len(maxBinary))
	}
}

// TestUploadSchema_ValidationError_InvalidVersionFormat tests that invalid version formats are rejected
func TestUploadSchema_ValidationError_InvalidVersionFormat(t *testing.T) {
	mockRepo := &mockSchemaRepository{}
	handler := NewSchemaHandler(mockRepo)
	client, cleanup := newTestClient(t, handler)
	defer cleanup()

	testCases := []struct {
		name    string
		version string
	}{
		{"missing patch", "1.0"},
		{"non-numeric", "v1.0.0"},
		{"with prefix", "version-1.0.0"},
		{"empty", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := connect.NewRequest(&isrv1.UploadSchemaRequest{
				Version:      tc.version,
				SchemaBinary: []byte("test"),
			})

			_, err := client.UploadSchema(context.Background(), req)
			if err == nil {
				t.Fatalf("UploadSchema() with invalid version %q should fail, but got nil error", tc.version)
			}

			var connectErr *connect.Error
			if !errors.As(err, &connectErr) {
				t.Fatalf("error type = %T, want *connect.Error", err)
			}

			if connectErr.Code() != connect.CodeInvalidArgument {
				t.Errorf("error code = %v, want %v (InvalidArgument)", connectErr.Code(), connect.CodeInvalidArgument)
			}
		})
	}
}

// TestGetLatestPatch_ValidationError_NegativeValues tests that negative major/minor values are rejected
func TestGetLatestPatch_ValidationError_NegativeValues(t *testing.T) {
	mockRepo := &mockSchemaRepository{}
	handler := NewSchemaHandler(mockRepo)
	client, cleanup := newTestClient(t, handler)
	defer cleanup()

	testCases := []struct {
		name  string
		major int32
		minor int32
	}{
		{"negative major", -1, 0},
		{"negative minor", 0, -1},
		{"both negative", -1, -1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := connect.NewRequest(&isrv1.GetLatestPatchRequest{
				Major: tc.major,
				Minor: tc.minor,
			})

			_, err := client.GetLatestPatch(context.Background(), req)
			if err == nil {
				t.Fatalf("GetLatestPatch() with major=%d, minor=%d should fail, but got nil error", tc.major, tc.minor)
			}

			var connectErr *connect.Error
			if !errors.As(err, &connectErr) {
				t.Fatalf("error type = %T, want *connect.Error", err)
			}

			if connectErr.Code() != connect.CodeInvalidArgument {
				t.Errorf("error code = %v, want %v (InvalidArgument)", connectErr.Code(), connect.CodeInvalidArgument)
			}
		})
	}
}

// TestGetSchemaByVersion_ValidationError_InvalidVersionFormat tests that invalid version formats are rejected
func TestGetSchemaByVersion_ValidationError_InvalidVersionFormat(t *testing.T) {
	mockRepo := &mockSchemaRepository{}
	handler := NewSchemaHandler(mockRepo)
	client, cleanup := newTestClient(t, handler)
	defer cleanup()

	testCases := []struct {
		name    string
		version string
	}{
		{"missing patch", "1.0"},
		{"non-numeric", "v1.0.0"},
		{"with prefix", "version-1.0.0"},
		{"empty", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := connect.NewRequest(&isrv1.GetSchemaByVersionRequest{
				Version: tc.version,
			})

			_, err := client.GetSchemaByVersion(context.Background(), req)
			if err == nil {
				t.Fatalf("GetSchemaByVersion() with invalid version %q should fail, but got nil error", tc.version)
			}

			var connectErr *connect.Error
			if !errors.As(err, &connectErr) {
				t.Fatalf("error type = %T, want *connect.Error", err)
			}

			if connectErr.Code() != connect.CodeInvalidArgument {
				t.Errorf("error code = %v, want %v (InvalidArgument)", connectErr.Code(), connect.CodeInvalidArgument)
			}
		})
	}
}
