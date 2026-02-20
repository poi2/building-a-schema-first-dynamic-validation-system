package handler

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	isrv1 "github.com/poi2/building-a-schema-first-dynamic-validation-system/pkg/gen/go/isr/v1"
	"github.com/poi2/building-a-schema-first-dynamic-validation-system/services/isr/internal/model"
	"github.com/poi2/building-a-schema-first-dynamic-validation-system/services/isr/internal/repository"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type SchemaHandler struct {
	repo *repository.SchemaRepository
}

func NewSchemaHandler(repo *repository.SchemaRepository) *SchemaHandler {
	return &SchemaHandler{repo: repo}
}

func (h *SchemaHandler) UploadSchema(
	ctx context.Context,
	req *connect.Request[isrv1.UploadSchemaRequest],
) (*connect.Response[isrv1.UploadSchemaResponse], error) {
	// Check if version already exists
	exists, err := h.repo.VersionExists(ctx, req.Msg.Version)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to check version: %w", err))
	}
	if exists {
		return nil, connect.NewError(connect.CodeAlreadyExists, fmt.Errorf("version %s already exists", req.Msg.Version))
	}

	// Generate UUID v7
	id, err := uuid.NewV7()
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to generate UUID: %w", err))
	}

	now := time.Now()
	schema := &model.Schema{
		ID:           id.String(),
		Version:      req.Msg.Version,
		SchemaBinary: req.Msg.SchemaBinary,
		SizeBytes:    int32(len(req.Msg.SchemaBinary)),
		CreatedAt:    now,
	}

	if err := h.repo.Create(ctx, schema); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to store schema: %w", err))
	}

	resp := &isrv1.UploadSchemaResponse{
		Metadata: &isrv1.SchemaMetadata{
			Id:        schema.ID,
			Version:   schema.Version,
			CreatedAt: timestamppb.New(schema.CreatedAt),
			SizeBytes: schema.SizeBytes,
		},
	}

	return connect.NewResponse(resp), nil
}

func (h *SchemaHandler) GetLatestPatch(
	ctx context.Context,
	req *connect.Request[isrv1.GetLatestPatchRequest],
) (*connect.Response[isrv1.GetLatestPatchResponse], error) {
	schema, err := h.repo.GetLatestPatch(ctx, req.Msg.Major, req.Msg.Minor)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("schema not found: %w", err))
	}

	resp := &isrv1.GetLatestPatchResponse{
		Metadata: &isrv1.SchemaMetadata{
			Id:        schema.ID,
			Version:   schema.Version,
			CreatedAt: timestamppb.New(schema.CreatedAt),
			SizeBytes: schema.SizeBytes,
		},
		SchemaBinary: schema.SchemaBinary,
	}

	return connect.NewResponse(resp), nil
}

func (h *SchemaHandler) GetSchemaByVersion(
	ctx context.Context,
	req *connect.Request[isrv1.GetSchemaByVersionRequest],
) (*connect.Response[isrv1.GetSchemaByVersionResponse], error) {
	schema, err := h.repo.GetByVersion(ctx, req.Msg.Version)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("schema not found: %w", err))
	}

	resp := &isrv1.GetSchemaByVersionResponse{
		Metadata: &isrv1.SchemaMetadata{
			Id:        schema.ID,
			Version:   schema.Version,
			CreatedAt: timestamppb.New(schema.CreatedAt),
			SizeBytes: schema.SizeBytes,
		},
		SchemaBinary: schema.SchemaBinary,
	}

	return connect.NewResponse(resp), nil
}
