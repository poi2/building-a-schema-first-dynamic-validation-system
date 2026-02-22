package validator

import (
	"fmt"
	"sync/atomic"

	"buf.build/go/protovalidate"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// validatorWithVersion wraps a validator with its schema version
type validatorWithVersion struct {
	validator protovalidate.Validator
	version   string
}

// SchemaAwareValidator provides thread-safe schema hot-swapping for protovalidate
type SchemaAwareValidator struct {
	v atomic.Value // *validatorWithVersion
}

// NewSchemaAwareValidator creates a new schema-aware validator with the given descriptor bytes and version
func NewSchemaAwareValidator(descriptorBytes []byte, version string) (*SchemaAwareValidator, error) {
	validator := &SchemaAwareValidator{}
	if err := validator.UpdateSchema(descriptorBytes, version); err != nil {
		return nil, fmt.Errorf("failed to initialize validator: %w", err)
	}
	return validator, nil
}

// Validate validates a protobuf message using the current schema
func (s *SchemaAwareValidator) Validate(msg proto.Message, options ...protovalidate.ValidationOption) error {
	vwv := s.v.Load().(*validatorWithVersion)
	return vwv.validator.Validate(msg, options...)
}

// UpdateSchema atomically updates the validator with a new schema
func (s *SchemaAwareValidator) UpdateSchema(descriptorBytes []byte, version string) error {
	// 1. Unmarshal FileDescriptorSet
	fds := &descriptorpb.FileDescriptorSet{}
	if err := proto.Unmarshal(descriptorBytes, fds); err != nil {
		return fmt.Errorf("failed to unmarshal descriptor set: %w", err)
	}

	// 2. Create Files registry
	files, err := protodesc.NewFiles(fds)
	if err != nil {
		return fmt.Errorf("failed to create files registry: %w", err)
	}

	// 3. Collect all message descriptors
	var descriptors []protoreflect.MessageDescriptor
	files.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		messages := fd.Messages()
		for i := 0; i < messages.Len(); i++ {
			descriptors = append(descriptors, messages.Get(i))
		}
		return true
	})

	if len(descriptors) == 0 {
		return fmt.Errorf("no message descriptors found in schema")
	}

	// 4. Create protovalidate.Validator
	validator, err := protovalidate.New(
		protovalidate.WithMessageDescriptors(descriptors...),
	)
	if err != nil {
		return fmt.Errorf("failed to create validator: %w", err)
	}

	// 5. Store in atomic.Value
	s.v.Store(&validatorWithVersion{
		validator: validator,
		version:   version,
	})

	return nil
}

// GetCurrentVersion returns the current schema version
func (s *SchemaAwareValidator) GetCurrentVersion() string {
	vwv := s.v.Load().(*validatorWithVersion)
	if vwv == nil {
		return ""
	}
	return vwv.version
}

// Ensure SchemaAwareValidator implements the protovalidate.Validator interface
var _ protovalidate.Validator = (*SchemaAwareValidator)(nil)
