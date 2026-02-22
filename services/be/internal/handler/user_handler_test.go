package handler

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	commonv1 "github.com/poi2/building-a-schema-first-dynamic-validation-system/pkg/gen/go/common/v1"
	userv1 "github.com/poi2/building-a-schema-first-dynamic-validation-system/pkg/gen/go/user/v1"
	"github.com/poi2/building-a-schema-first-dynamic-validation-system/services/be/internal/model"
)

// Mock repository for testing
type mockUserRepository struct {
	users map[string]*model.User
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users: make(map[string]*model.User),
	}
}

func (m *mockUserRepository) Create(ctx context.Context, user *model.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepository) List(ctx context.Context, page, pageSize int) ([]*model.User, int, error) {
	var users []*model.User
	for _, user := range m.users {
		users = append(users, user)
	}

	// Simple pagination
	start := (page - 1) * pageSize
	end := start + pageSize
	total := len(users)

	if start >= total {
		return []*model.User{}, total, nil
	}
	if end > total {
		end = total
	}

	return users[start:end], total, nil
}

func (m *mockUserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	user, ok := m.users[id]
	if !ok {
		return nil, connect.NewError(connect.CodeNotFound, nil)
	}
	return user, nil
}

func TestUserHandler_CreateUser(t *testing.T) {
	repo := newMockUserRepository()
	handler := NewUserHandler(repo)
	ctx := context.Background()

	req := connect.NewRequest(&userv1.CreateUserRequest{
		Name:  "Test User",
		Email: "test@example.com",
		Plan:  commonv1.UserPlan_USER_PLAN_FREE,
	})

	resp, err := handler.CreateUser(ctx, req)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if resp.Msg.User.Name != "Test User" {
		t.Errorf("Expected name 'Test User', got '%s'", resp.Msg.User.Name)
	}
	if resp.Msg.User.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", resp.Msg.User.Email)
	}
	if resp.Msg.User.Plan != commonv1.UserPlan_USER_PLAN_FREE {
		t.Errorf("Expected plan FREE, got %v", resp.Msg.User.Plan)
	}
	if resp.Msg.User.Id == "" {
		t.Error("Expected non-empty user ID")
	}
}

func TestUserHandler_ListUsers(t *testing.T) {
	repo := newMockUserRepository()
	handler := NewUserHandler(repo)
	ctx := context.Background()

	// Create some users
	for i := 0; i < 5; i++ {
		repo.users["user-"+string(rune('0'+i))] = &model.User{
			ID:        "user-" + string(rune('0'+i)),
			Name:      "User " + string(rune('A'+i)),
			Email:     "user" + string(rune('0'+i)) + "@example.com",
			Plan:      "free",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	req := connect.NewRequest(&userv1.ListUsersRequest{
		Page:     1,
		PageSize: 3,
	})

	resp, err := handler.ListUsers(ctx, req)
	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}

	if resp.Msg.Total != 5 {
		t.Errorf("Expected total 5, got %d", resp.Msg.Total)
	}
	if len(resp.Msg.Users) != 3 {
		t.Errorf("Expected 3 users, got %d", len(resp.Msg.Users))
	}
}


func TestUserPlanConversion(t *testing.T) {
	tests := []struct {
		plan     commonv1.UserPlan
		expected string
	}{
		{commonv1.UserPlan_USER_PLAN_FREE, "free"},
		{commonv1.UserPlan_USER_PLAN_PRO, "pro"},
		{commonv1.UserPlan_USER_PLAN_ENTERPRISE, "enterprise"},
		{commonv1.UserPlan_USER_PLAN_UNSPECIFIED, "free"},
	}

	for _, tt := range tests {
		result := userPlanToString(tt.plan)
		if result != tt.expected {
			t.Errorf("userPlanToString(%v) = %s, want %s", tt.plan, result, tt.expected)
		}
	}
}

func TestStringToUserPlan(t *testing.T) {
	tests := []struct {
		str      string
		expected commonv1.UserPlan
	}{
		{"free", commonv1.UserPlan_USER_PLAN_FREE},
		{"pro", commonv1.UserPlan_USER_PLAN_PRO},
		{"enterprise", commonv1.UserPlan_USER_PLAN_ENTERPRISE},
		{"unknown", commonv1.UserPlan_USER_PLAN_FREE},
	}

	for _, tt := range tests {
		result := stringToUserPlan(tt.str)
		if result != tt.expected {
			t.Errorf("stringToUserPlan(%s) = %v, want %v", tt.str, result, tt.expected)
		}
	}
}
