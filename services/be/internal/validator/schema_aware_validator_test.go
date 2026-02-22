package validator

import (
	"sync"
	"testing"

	commonv1 "github.com/poi2/building-a-schema-first-dynamic-validation-system/pkg/gen/go/common/v1"
	userv1 "github.com/poi2/building-a-schema-first-dynamic-validation-system/pkg/gen/go/user/v1"
	"google.golang.org/protobuf/proto"
)

func TestNewSchemaAwareValidator_Success(t *testing.T) {
	descriptorBytes := CreateTestDescriptorBytes(t)

	validator, err := NewSchemaAwareValidator(descriptorBytes, "1.0.0")
	if err != nil {
		t.Fatalf("NewSchemaAwareValidator failed: %v", err)
	}

	if validator == nil {
		t.Fatal("validator is nil")
	}

	version := validator.GetCurrentVersion()
	if version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", version)
	}
}

func TestNewSchemaAwareValidator_InvalidDescriptor(t *testing.T) {
	invalidBytes := []byte("invalid descriptor data")

	_, err := NewSchemaAwareValidator(invalidBytes, "1.0.0")
	if err == nil {
		t.Fatal("expected error for invalid descriptor, got nil")
	}
}

func TestSchemaAwareValidator_Validate(t *testing.T) {
	descriptorBytes := CreateTestDescriptorBytes(t)

	validator, err := NewSchemaAwareValidator(descriptorBytes, "1.0.0")
	if err != nil {
		t.Fatalf("NewSchemaAwareValidator failed: %v", err)
	}

	tests := []struct {
		name    string
		msg     proto.Message
		wantErr bool
	}{
		{
			name: "valid CreateUserRequest",
			msg: &userv1.CreateUserRequest{
				Name:  "John Doe",
				Email: "john@example.com",
				Plan:  commonv1.UserPlan_USER_PLAN_FREE,
			},
			wantErr: false,
		},
		{
			name: "invalid CreateUserRequest - empty name",
			msg: &userv1.CreateUserRequest{
				Name:  "",
				Email: "john@example.com",
				Plan:  commonv1.UserPlan_USER_PLAN_FREE,
			},
			wantErr: true,
		},
		{
			name: "invalid CreateUserRequest - invalid email",
			msg: &userv1.CreateUserRequest{
				Name:  "John Doe",
				Email: "not-an-email",
				Plan:  commonv1.UserPlan_USER_PLAN_FREE,
			},
			wantErr: true,
		},
		{
			name: "valid ListUsersRequest",
			msg: &userv1.ListUsersRequest{
				Page:     1,
				PageSize: 10,
			},
			wantErr: false,
		},
		{
			name: "invalid ListUsersRequest - page too small",
			msg: &userv1.ListUsersRequest{
				Page:     0,
				PageSize: 10,
			},
			wantErr: true,
		},
		{
			name: "invalid ListUsersRequest - page_size too large",
			msg: &userv1.ListUsersRequest{
				Page:     1,
				PageSize: 101,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSchemaAwareValidator_UpdateSchema(t *testing.T) {
	descriptorBytes := CreateTestDescriptorBytes(t)

	validator, err := NewSchemaAwareValidator(descriptorBytes, "1.0.0")
	if err != nil {
		t.Fatalf("NewSchemaAwareValidator failed: %v", err)
	}

	initialVersion := validator.GetCurrentVersion()
	if initialVersion != "1.0.0" {
		t.Errorf("expected initial version 1.0.0, got %s", initialVersion)
	}

	// Update to new version
	err = validator.UpdateSchema(descriptorBytes, "1.0.1")
	if err != nil {
		t.Fatalf("UpdateSchema failed: %v", err)
	}

	newVersion := validator.GetCurrentVersion()
	if newVersion != "1.0.1" {
		t.Errorf("expected new version 1.0.1, got %s", newVersion)
	}

	// Verify validator still works after update
	validMsg := &userv1.CreateUserRequest{
		Name:  "John Doe",
		Email: "john@example.com",
		Plan:  commonv1.UserPlan_USER_PLAN_FREE,
	}
	if err := validator.Validate(validMsg); err != nil {
		t.Errorf("Validate failed after update: %v", err)
	}
}

func TestSchemaAwareValidator_ConcurrentAccess(t *testing.T) {
	descriptorBytes := CreateTestDescriptorBytes(t)

	validator, err := NewSchemaAwareValidator(descriptorBytes, "1.0.0")
	if err != nil {
		t.Fatalf("NewSchemaAwareValidator failed: %v", err)
	}

	// Run concurrent reads and writes
	var wg sync.WaitGroup
	numReaders := 10
	numWriters := 5
	iterations := 100

	// Start readers
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				msg := &userv1.CreateUserRequest{
					Name:  "John Doe",
					Email: "john@example.com",
					Plan:  commonv1.UserPlan_USER_PLAN_FREE,
				}
				_ = validator.Validate(msg)
				_ = validator.GetCurrentVersion()
			}
		}(i)
	}

	// Start writers
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				version := "1.0." + string(rune('0'+j%10))
				_ = validator.UpdateSchema(descriptorBytes, version)
			}
		}(i)
	}

	wg.Wait()

	// Verify validator is still functional
	validMsg := &userv1.CreateUserRequest{
		Name:  "John Doe",
		Email: "john@example.com",
		Plan:  commonv1.UserPlan_USER_PLAN_FREE,
	}
	if err := validator.Validate(validMsg); err != nil {
		t.Errorf("Validate failed after concurrent access: %v", err)
	}
}

func TestSchemaAwareValidator_GetCurrentVersion(t *testing.T) {
	descriptorBytes := CreateTestDescriptorBytes(t)

	tests := []struct {
		name    string
		version string
	}{
		{
			name:    "version 1.0.0",
			version: "1.0.0",
		},
		{
			name:    "version 1.0.1",
			version: "1.0.1",
		},
		{
			name:    "version 2.0.0",
			version: "2.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator, err := NewSchemaAwareValidator(descriptorBytes, tt.version)
			if err != nil {
				t.Fatalf("NewSchemaAwareValidator failed: %v", err)
			}

			got := validator.GetCurrentVersion()
			if got != tt.version {
				t.Errorf("GetCurrentVersion() = %v, want %v", got, tt.version)
			}
		})
	}
}
