package handler

import (
	"context"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	isrv1 "github.com/poi2/building-a-schema-first-dynamic-validation-system/pkg/gen/go/isr/v1"
	"github.com/poi2/building-a-schema-first-dynamic-validation-system/services/isr/internal/model"
)

// mockSchemaRepository is a mock implementation of SchemaRepositoryInterface
type mockSchemaRepository struct {
	createFunc         func(ctx context.Context, schema *model.Schema) error
	getByVersionFunc   func(ctx context.Context, version string) (*model.Schema, error)
	getLatestPatchFunc func(ctx context.Context, major, minor int32) (*model.Schema, error)
	versionExistsFunc  func(ctx context.Context, version string) (bool, error)
}

func (m *mockSchemaRepository) Create(ctx context.Context, schema *model.Schema) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, schema)
	}
	return nil
}

func (m *mockSchemaRepository) GetByVersion(ctx context.Context, version string) (*model.Schema, error) {
	if m.getByVersionFunc != nil {
		return m.getByVersionFunc(ctx, version)
	}
	return nil, nil
}

func (m *mockSchemaRepository) GetLatestPatch(ctx context.Context, major, minor int32) (*model.Schema, error) {
	if m.getLatestPatchFunc != nil {
		return m.getLatestPatchFunc(ctx, major, minor)
	}
	return nil, nil
}

func (m *mockSchemaRepository) VersionExists(ctx context.Context, version string) (bool, error) {
	if m.versionExistsFunc != nil {
		return m.versionExistsFunc(ctx, version)
	}
	return false, nil
}

func TestSchemaHandler_UploadSchema_Success(t *testing.T) {
	mockRepo := &mockSchemaRepository{
		versionExistsFunc: func(ctx context.Context, version string) (bool, error) {
			return false, nil
		},
		createFunc: func(ctx context.Context, schema *model.Schema) error {
			return nil
		},
	}

	handler := NewSchemaHandler(mockRepo)

	req := connect.NewRequest(&isrv1.UploadSchemaRequest{
		Version:      "1.2.3",
		SchemaBinary: []byte("test schema data"),
	})

	resp, err := handler.UploadSchema(context.Background(), req)
	if err != nil {
		t.Fatalf("UploadSchema() error = %v, want nil", err)
	}

	if resp.Msg.Metadata.Version != "1.2.3" {
		t.Errorf("Version = %v, want 1.2.3", resp.Msg.Metadata.Version)
	}
	if resp.Msg.Metadata.SizeBytes != int32(len("test schema data")) {
		t.Errorf("SizeBytes = %v, want %v", resp.Msg.Metadata.SizeBytes, len("test schema data"))
	}
	if resp.Msg.Metadata.Id == "" {
		t.Errorf("Id is empty, want non-empty value")
	}
	if resp.Msg.Metadata.CreatedAt == nil {
		t.Errorf("CreatedAt is nil, want non-nil value")
	} else if resp.Msg.Metadata.CreatedAt.AsTime().IsZero() {
		t.Errorf("CreatedAt is zero, want non-zero timestamp")
	}
}

func TestSchemaHandler_UploadSchema_InvalidVersion(t *testing.T) {
	mockRepo := &mockSchemaRepository{}
	handler := NewSchemaHandler(mockRepo)

	req := connect.NewRequest(&isrv1.UploadSchemaRequest{
		Version:      "invalid-version",
		SchemaBinary: []byte("test"),
	})

	_, err := handler.UploadSchema(context.Background(), req)
	if err == nil {
		t.Fatal("UploadSchema() error = nil, want error")
	}

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("error type = %T, want *connect.Error", err)
	}

	if connectErr.Code() != connect.CodeInvalidArgument {
		t.Errorf("error code = %v, want %v", connectErr.Code(), connect.CodeInvalidArgument)
	}
}

func TestSchemaHandler_UploadSchema_AlreadyExists(t *testing.T) {
	mockRepo := &mockSchemaRepository{
		versionExistsFunc: func(ctx context.Context, version string) (bool, error) {
			return true, nil
		},
	}

	handler := NewSchemaHandler(mockRepo)

	req := connect.NewRequest(&isrv1.UploadSchemaRequest{
		Version:      "1.2.3",
		SchemaBinary: []byte("test"),
	})

	_, err := handler.UploadSchema(context.Background(), req)
	if err == nil {
		t.Fatal("UploadSchema() error = nil, want error")
	}

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("error type = %T, want *connect.Error", err)
	}

	if connectErr.Code() != connect.CodeAlreadyExists {
		t.Errorf("error code = %v, want %v", connectErr.Code(), connect.CodeAlreadyExists)
	}
}

func TestSchemaHandler_UploadSchema_RepositoryError(t *testing.T) {
	mockRepo := &mockSchemaRepository{
		versionExistsFunc: func(ctx context.Context, version string) (bool, error) {
			return false, errors.New("database error")
		},
	}

	handler := NewSchemaHandler(mockRepo)

	req := connect.NewRequest(&isrv1.UploadSchemaRequest{
		Version:      "1.2.3",
		SchemaBinary: []byte("test"),
	})

	_, err := handler.UploadSchema(context.Background(), req)
	if err == nil {
		t.Fatal("UploadSchema() error = nil, want error")
	}

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("error type = %T, want *connect.Error", err)
	}

	if connectErr.Code() != connect.CodeInternal {
		t.Errorf("error code = %v, want %v", connectErr.Code(), connect.CodeInternal)
	}
}

func TestSchemaHandler_UploadSchema_CreateError(t *testing.T) {
	mockRepo := &mockSchemaRepository{
		versionExistsFunc: func(ctx context.Context, version string) (bool, error) {
			return false, nil
		},
		createFunc: func(ctx context.Context, schema *model.Schema) error {
			return errors.New("database error")
		},
	}

	handler := NewSchemaHandler(mockRepo)

	req := connect.NewRequest(&isrv1.UploadSchemaRequest{
		Version:      "1.2.3",
		SchemaBinary: []byte("test"),
	})

	_, err := handler.UploadSchema(context.Background(), req)
	if err == nil {
		t.Fatal("UploadSchema() error = nil, want error")
	}

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("error type = %T, want *connect.Error", err)
	}

	if connectErr.Code() != connect.CodeInternal {
		t.Errorf("error code = %v, want %v", connectErr.Code(), connect.CodeInternal)
	}
}

func TestSchemaHandler_GetLatestPatch_Success(t *testing.T) {
	expectedSchema := &model.Schema{
		ID:           "test-id",
		Version:      "1.2.5",
		Major:        1,
		Minor:        2,
		Patch:        5,
		SchemaBinary: []byte("test schema"),
		SizeBytes:    11,
		CreatedAt:    time.Now(),
	}

	mockRepo := &mockSchemaRepository{
		getLatestPatchFunc: func(ctx context.Context, major, minor int32) (*model.Schema, error) {
			if major == 1 && minor == 2 {
				return expectedSchema, nil
			}
			return nil, pgx.ErrNoRows
		},
	}

	handler := NewSchemaHandler(mockRepo)

	req := connect.NewRequest(&isrv1.GetLatestPatchRequest{
		Major: 1,
		Minor: 2,
	})

	resp, err := handler.GetLatestPatch(context.Background(), req)
	if err != nil {
		t.Fatalf("GetLatestPatch() error = %v, want nil", err)
	}

	if resp.Msg.Metadata.Version != "1.2.5" {
		t.Errorf("Version = %v, want 1.2.5", resp.Msg.Metadata.Version)
	}
	if string(resp.Msg.SchemaBinary) != "test schema" {
		t.Errorf("SchemaBinary = %v, want 'test schema'", string(resp.Msg.SchemaBinary))
	}
	if resp.Msg.Metadata.Id != "test-id" {
		t.Errorf("Id = %v, want test-id", resp.Msg.Metadata.Id)
	}
	if resp.Msg.Metadata.SizeBytes != 11 {
		t.Errorf("SizeBytes = %v, want 11", resp.Msg.Metadata.SizeBytes)
	}
	if resp.Msg.Metadata.CreatedAt == nil {
		t.Errorf("CreatedAt is nil, want non-nil value")
	}
}

func TestSchemaHandler_GetLatestPatch_NotFound(t *testing.T) {
	mockRepo := &mockSchemaRepository{
		getLatestPatchFunc: func(ctx context.Context, major, minor int32) (*model.Schema, error) {
			return nil, pgx.ErrNoRows
		},
	}

	handler := NewSchemaHandler(mockRepo)

	req := connect.NewRequest(&isrv1.GetLatestPatchRequest{
		Major: 99,
		Minor: 99,
	})

	_, err := handler.GetLatestPatch(context.Background(), req)
	if err == nil {
		t.Fatal("GetLatestPatch() error = nil, want error")
	}

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("error type = %T, want *connect.Error", err)
	}

	if connectErr.Code() != connect.CodeNotFound {
		t.Errorf("error code = %v, want %v", connectErr.Code(), connect.CodeNotFound)
	}
}

func TestSchemaHandler_GetLatestPatch_InternalError(t *testing.T) {
	mockRepo := &mockSchemaRepository{
		getLatestPatchFunc: func(ctx context.Context, major, minor int32) (*model.Schema, error) {
			return nil, errors.New("database connection failed")
		},
	}

	handler := NewSchemaHandler(mockRepo)

	req := connect.NewRequest(&isrv1.GetLatestPatchRequest{
		Major: 1,
		Minor: 0,
	})

	_, err := handler.GetLatestPatch(context.Background(), req)
	if err == nil {
		t.Fatal("GetLatestPatch() error = nil, want error")
	}

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("error type = %T, want *connect.Error", err)
	}

	if connectErr.Code() != connect.CodeInternal {
		t.Errorf("error code = %v, want %v", connectErr.Code(), connect.CodeInternal)
	}
}

func TestSchemaHandler_GetSchemaByVersion_Success(t *testing.T) {
	expectedSchema := &model.Schema{
		ID:           "test-id",
		Version:      "2.0.1",
		Major:        2,
		Minor:        0,
		Patch:        1,
		SchemaBinary: []byte("version 2.0.1 schema"),
		SizeBytes:    20,
		CreatedAt:    time.Now(),
	}

	mockRepo := &mockSchemaRepository{
		getByVersionFunc: func(ctx context.Context, version string) (*model.Schema, error) {
			if version == "2.0.1" {
				return expectedSchema, nil
			}
			return nil, pgx.ErrNoRows
		},
	}

	handler := NewSchemaHandler(mockRepo)

	req := connect.NewRequest(&isrv1.GetSchemaByVersionRequest{
		Version: "2.0.1",
	})

	resp, err := handler.GetSchemaByVersion(context.Background(), req)
	if err != nil {
		t.Fatalf("GetSchemaByVersion() error = %v, want nil", err)
	}

	if resp.Msg.Metadata.Version != "2.0.1" {
		t.Errorf("Version = %v, want 2.0.1", resp.Msg.Metadata.Version)
	}
	if string(resp.Msg.SchemaBinary) != "version 2.0.1 schema" {
		t.Errorf("SchemaBinary = %v, want 'version 2.0.1 schema'", string(resp.Msg.SchemaBinary))
	}
	if resp.Msg.Metadata.Id != "test-id" {
		t.Errorf("Id = %v, want test-id", resp.Msg.Metadata.Id)
	}
	if resp.Msg.Metadata.SizeBytes != 20 {
		t.Errorf("SizeBytes = %v, want 20", resp.Msg.Metadata.SizeBytes)
	}
	if resp.Msg.Metadata.CreatedAt == nil {
		t.Errorf("CreatedAt is nil, want non-nil value")
	}
}

func TestSchemaHandler_GetSchemaByVersion_NotFound(t *testing.T) {
	mockRepo := &mockSchemaRepository{
		getByVersionFunc: func(ctx context.Context, version string) (*model.Schema, error) {
			return nil, pgx.ErrNoRows
		},
	}

	handler := NewSchemaHandler(mockRepo)

	req := connect.NewRequest(&isrv1.GetSchemaByVersionRequest{
		Version: "99.99.99",
	})

	_, err := handler.GetSchemaByVersion(context.Background(), req)
	if err == nil {
		t.Fatal("GetSchemaByVersion() error = nil, want error")
	}

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("error type = %T, want *connect.Error", err)
	}

	if connectErr.Code() != connect.CodeNotFound {
		t.Errorf("error code = %v, want %v", connectErr.Code(), connect.CodeNotFound)
	}
}

func TestSchemaHandler_GetSchemaByVersion_InternalError(t *testing.T) {
	mockRepo := &mockSchemaRepository{
		getByVersionFunc: func(ctx context.Context, version string) (*model.Schema, error) {
			return nil, errors.New("database connection failed")
		},
	}

	handler := NewSchemaHandler(mockRepo)

	req := connect.NewRequest(&isrv1.GetSchemaByVersionRequest{
		Version: "1.0.0",
	})

	_, err := handler.GetSchemaByVersion(context.Background(), req)
	if err == nil {
		t.Fatal("GetSchemaByVersion() error = nil, want error")
	}

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("error type = %T, want *connect.Error", err)
	}

	if connectErr.Code() != connect.CodeInternal {
		t.Errorf("error code = %v, want %v", connectErr.Code(), connect.CodeInternal)
	}
}
