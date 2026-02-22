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
	commonv1 "github.com/poi2/building-a-schema-first-dynamic-validation-system/pkg/gen/go/common/v1"
	userv1 "github.com/poi2/building-a-schema-first-dynamic-validation-system/pkg/gen/go/user/v1"
	userv1connect "github.com/poi2/building-a-schema-first-dynamic-validation-system/pkg/gen/go/user/v1/userv1connect"
)

// newTestClient creates a test Connect client with validation interceptor
func newTestClient(t *testing.T, handler *UserHandler) (userv1connect.UserServiceClient, func()) {
	t.Helper()

	mux := http.NewServeMux()
	interceptors := connect.WithInterceptors(
		validate.NewInterceptor(),
	)
	path, connectHandler := userv1connect.NewUserServiceHandler(handler, interceptors)
	mux.Handle(path, connectHandler)

	server := httptest.NewServer(mux)
	client := userv1connect.NewUserServiceClient(
		http.DefaultClient,
		server.URL,
	)

	return client, server.Close
}

func TestCreateUser_ValidationError_EmptyName(t *testing.T) {
	repo := newMockUserRepository()
	handler := NewUserHandler(repo)
	client, cleanup := newTestClient(t, handler)
	defer cleanup()

	req := connect.NewRequest(&userv1.CreateUserRequest{
		Name:  "", // Empty name
		Email: "test@example.com",
		Plan:  commonv1.UserPlan_USER_PLAN_FREE,
	})

	_, err := client.CreateUser(context.Background(), req)
	if err == nil {
		t.Fatal("CreateUser() with empty name should fail, but got nil error")
	}

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("error type = %T, want *connect.Error", err)
	}
	if connectErr.Code() != connect.CodeInvalidArgument {
		t.Errorf("error code = %v, want %v (InvalidArgument)", connectErr.Code(), connect.CodeInvalidArgument)
	}
	if !strings.Contains(connectErr.Message(), "value length must be at least 1 characters") {
		t.Errorf("error message = %q, want to contain 'value length must be at least 1 characters'", connectErr.Message())
	}
}

func TestCreateUser_ValidationError_InvalidEmail(t *testing.T) {
	repo := newMockUserRepository()
	handler := NewUserHandler(repo)
	client, cleanup := newTestClient(t, handler)
	defer cleanup()

	req := connect.NewRequest(&userv1.CreateUserRequest{
		Name:  "Test User",
		Email: "invalid-email", // Invalid email
		Plan:  commonv1.UserPlan_USER_PLAN_FREE,
	})

	_, err := client.CreateUser(context.Background(), req)
	if err == nil {
		t.Fatal("CreateUser() with invalid email should fail, but got nil error")
	}

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("error type = %T, want *connect.Error", err)
	}
	if connectErr.Code() != connect.CodeInvalidArgument {
		t.Errorf("error code = %v, want %v (InvalidArgument)", connectErr.Code(), connect.CodeInvalidArgument)
	}
	if !strings.Contains(connectErr.Message(), "value must be a valid email address") {
		t.Errorf("error message = %q, want to contain 'value must be a valid email address'", connectErr.Message())
	}
}

func TestCreateUser_ValidationError_NameTooLong(t *testing.T) {
	repo := newMockUserRepository()
	handler := NewUserHandler(repo)
	client, cleanup := newTestClient(t, handler)
	defer cleanup()

	req := connect.NewRequest(&userv1.CreateUserRequest{
		Name:  strings.Repeat("a", 101), // Name too long (max 100)
		Email: "test@example.com",
		Plan:  commonv1.UserPlan_USER_PLAN_FREE,
	})

	_, err := client.CreateUser(context.Background(), req)
	if err == nil {
		t.Fatal("CreateUser() with name too long should fail, but got nil error")
	}

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("error type = %T, want *connect.Error", err)
	}
	if connectErr.Code() != connect.CodeInvalidArgument {
		t.Errorf("error code = %v, want %v (InvalidArgument)", connectErr.Code(), connect.CodeInvalidArgument)
	}
	if !strings.Contains(connectErr.Message(), "value length must be at most 100 characters") {
		t.Errorf("error message = %q, want to contain 'value length must be at most 100 characters'", connectErr.Message())
	}
}

func TestListUsers_ValidationError_PageLessThan1(t *testing.T) {
	repo := newMockUserRepository()
	handler := NewUserHandler(repo)
	client, cleanup := newTestClient(t, handler)
	defer cleanup()

	req := connect.NewRequest(&userv1.ListUsersRequest{
		Page:     0, // Page < 1
		PageSize: 10,
	})

	_, err := client.ListUsers(context.Background(), req)
	if err == nil {
		t.Fatal("ListUsers() with page < 1 should fail, but got nil error")
	}

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("error type = %T, want *connect.Error", err)
	}
	if connectErr.Code() != connect.CodeInvalidArgument {
		t.Errorf("error code = %v, want %v (InvalidArgument)", connectErr.Code(), connect.CodeInvalidArgument)
	}
	if !strings.Contains(connectErr.Message(), "value must be greater than or equal to 1") {
		t.Errorf("error message = %q, want to contain 'value must be greater than or equal to 1'", connectErr.Message())
	}
}

func TestListUsers_ValidationError_PageSizeLessThan1(t *testing.T) {
	repo := newMockUserRepository()
	handler := NewUserHandler(repo)
	client, cleanup := newTestClient(t, handler)
	defer cleanup()

	req := connect.NewRequest(&userv1.ListUsersRequest{
		Page:     1,
		PageSize: 0, // PageSize < 1
	})

	_, err := client.ListUsers(context.Background(), req)
	if err == nil {
		t.Fatal("ListUsers() with pageSize < 1 should fail, but got nil error")
	}

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("error type = %T, want *connect.Error", err)
	}
	if connectErr.Code() != connect.CodeInvalidArgument {
		t.Errorf("error code = %v, want %v (InvalidArgument)", connectErr.Code(), connect.CodeInvalidArgument)
	}
	if !strings.Contains(connectErr.Message(), "value must be greater than or equal to 1") {
		t.Errorf("error message = %q, want to contain 'value must be greater than or equal to 1'", connectErr.Message())
	}
}

func TestListUsers_ValidationError_PageSizeGreaterThan100(t *testing.T) {
	repo := newMockUserRepository()
	handler := NewUserHandler(repo)
	client, cleanup := newTestClient(t, handler)
	defer cleanup()

	req := connect.NewRequest(&userv1.ListUsersRequest{
		Page:     1,
		PageSize: 101, // PageSize > 100
	})

	_, err := client.ListUsers(context.Background(), req)
	if err == nil {
		t.Fatal("ListUsers() with pageSize > 100 should fail, but got nil error")
	}

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("error type = %T, want *connect.Error", err)
	}
	if connectErr.Code() != connect.CodeInvalidArgument {
		t.Errorf("error code = %v, want %v (InvalidArgument)", connectErr.Code(), connect.CodeInvalidArgument)
	}
	if !strings.Contains(connectErr.Message(), "less than or equal to 100") {
		t.Errorf("error message = %q, want to contain 'less than or equal to 100'", connectErr.Message())
	}
}
