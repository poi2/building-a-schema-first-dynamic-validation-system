package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"connectrpc.com/validate"
	isrv1 "github.com/poi2/building-a-schema-first-dynamic-validation-system/pkg/gen/go/isr/v1"
	"github.com/poi2/building-a-schema-first-dynamic-validation-system/pkg/gen/go/isr/v1/isrv1connect"
	"github.com/poi2/building-a-schema-first-dynamic-validation-system/services/isr/internal/model"
)

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

	// Create server with validation interceptor
	mux := http.NewServeMux()
	interceptors := connect.WithInterceptors(
		validate.NewInterceptor(),
	)
	path, connectHandler := isrv1connect.NewSchemaRegistryServiceHandler(handler, interceptors)
	mux.Handle(path, connectHandler)

	server := httptest.NewServer(mux)
	defer server.Close()

	client := isrv1connect.NewSchemaRegistryServiceClient(
		http.DefaultClient,
		server.URL,
	)

	// Create a binary larger than 10MB (10485760 bytes)
	largeBinary := make([]byte, 10485761) // 10MB + 1 byte
	for i := range largeBinary {
		largeBinary[i] = byte(i % 256)
	}

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

	// Verify error message mentions validation
	if !strings.Contains(connectErr.Message(), "validation") && !strings.Contains(connectErr.Message(), "max_len") {
		t.Errorf("error message should mention validation or max_len, got: %s", connectErr.Message())
	}
}

// TestUploadSchema_ValidationError_SchemaBinaryEmpty tests that empty schema binaries are rejected
func TestUploadSchema_ValidationError_SchemaBinaryEmpty(t *testing.T) {
	mockRepo := &mockSchemaRepository{}
	handler := NewSchemaHandler(mockRepo)

	// Create server with validation interceptor
	mux := http.NewServeMux()
	interceptors := connect.WithInterceptors(
		validate.NewInterceptor(),
	)
	path, connectHandler := isrv1connect.NewSchemaRegistryServiceHandler(handler, interceptors)
	mux.Handle(path, connectHandler)

	server := httptest.NewServer(mux)
	defer server.Close()

	client := isrv1connect.NewSchemaRegistryServiceClient(
		http.DefaultClient,
		server.URL,
	)

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

	// Create server with validation interceptor
	mux := http.NewServeMux()
	interceptors := connect.WithInterceptors(
		validate.NewInterceptor(),
	)
	path, connectHandler := isrv1connect.NewSchemaRegistryServiceHandler(handler, interceptors)
	mux.Handle(path, connectHandler)

	server := httptest.NewServer(mux)
	defer server.Close()

	client := isrv1connect.NewSchemaRegistryServiceClient(
		http.DefaultClient,
		server.URL,
	)

	// Create a binary exactly 10MB (10485760 bytes)
	maxBinary := make([]byte, 10485760)
	for i := range maxBinary {
		maxBinary[i] = byte(i % 256)
	}

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

	// Create server with validation interceptor
	mux := http.NewServeMux()
	interceptors := connect.WithInterceptors(
		validate.NewInterceptor(),
	)
	path, connectHandler := isrv1connect.NewSchemaRegistryServiceHandler(handler, interceptors)
	mux.Handle(path, connectHandler)

	server := httptest.NewServer(mux)
	defer server.Close()

	client := isrv1connect.NewSchemaRegistryServiceClient(
		http.DefaultClient,
		server.URL,
	)

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

	// Create server with validation interceptor
	mux := http.NewServeMux()
	interceptors := connect.WithInterceptors(
		validate.NewInterceptor(),
	)
	path, connectHandler := isrv1connect.NewSchemaRegistryServiceHandler(handler, interceptors)
	mux.Handle(path, connectHandler)

	server := httptest.NewServer(mux)
	defer server.Close()

	client := isrv1connect.NewSchemaRegistryServiceClient(
		http.DefaultClient,
		server.URL,
	)

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

	// Create server with validation interceptor
	mux := http.NewServeMux()
	interceptors := connect.WithInterceptors(
		validate.NewInterceptor(),
	)
	path, connectHandler := isrv1connect.NewSchemaRegistryServiceHandler(handler, interceptors)
	mux.Handle(path, connectHandler)

	server := httptest.NewServer(mux)
	defer server.Close()

	client := isrv1connect.NewSchemaRegistryServiceClient(
		http.DefaultClient,
		server.URL,
	)

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
