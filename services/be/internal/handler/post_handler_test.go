package handler

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"connectrpc.com/connect"
	commonv1 "github.com/poi2/building-a-schema-first-dynamic-validation-system/pkg/gen/go/common/v1"
	postv1 "github.com/poi2/building-a-schema-first-dynamic-validation-system/pkg/gen/go/post/v1"
	"github.com/poi2/building-a-schema-first-dynamic-validation-system/services/be/internal/model"
)

// Mock post repository for testing
type mockPostRepository struct {
	posts map[string]*model.Post
}

func newMockPostRepository() *mockPostRepository {
	return &mockPostRepository{
		posts: make(map[string]*model.Post),
	}
}

func (m *mockPostRepository) Create(ctx context.Context, post *model.Post) error {
	m.posts[post.ID] = post
	return nil
}

func (m *mockPostRepository) List(ctx context.Context, userID string, page, pageSize int) ([]*model.Post, int, error) {
	var userPosts []*model.Post
	for _, post := range m.posts {
		if post.UserID == userID {
			userPosts = append(userPosts, post)
		}
	}

	total := len(userPosts)
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= total {
		return []*model.Post{}, total, nil
	}
	if end > total {
		end = total
	}

	return userPosts[start:end], total, nil
}

func (m *mockPostRepository) GetByID(ctx context.Context, id string) (*model.Post, error) {
	post, ok := m.posts[id]
	if !ok {
		return nil, os.ErrNotExist
	}
	return post, nil
}

// Mock user repository that returns users with different plans
type mockUserRepositoryForPost struct {
	users map[string]*model.User
}

func newMockUserRepositoryForPost() *mockUserRepositoryForPost {
	return &mockUserRepositoryForPost{
		users: make(map[string]*model.User),
	}
}

func (m *mockUserRepositoryForPost) Create(ctx context.Context, user *model.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepositoryForPost) List(ctx context.Context, page, pageSize int) ([]*model.User, int, error) {
	return nil, 0, nil
}

func (m *mockUserRepositoryForPost) GetByID(ctx context.Context, id string) (*model.User, error) {
	user, ok := m.users[id]
	if !ok {
		return nil, os.ErrNotExist
	}
	return user, nil
}

func TestPostHandler_CreatePost_PlanBasedContentLimit(t *testing.T) {
	postRepo := newMockPostRepository()
	userRepo := newMockUserRepositoryForPost()
	handler := NewPostHandler(postRepo, userRepo)
	ctx := context.Background()

	tests := []struct {
		name          string
		userPlan      string
		contentLength int
		shouldFail    bool
	}{
		{"FREE plan - within limit", "free", 500, false},
		{"FREE plan - at limit", "free", 1000, false},
		{"FREE plan - exceeds limit", "free", 1001, true},
		{"PRO plan - within limit", "pro", 3000, false},
		{"PRO plan - at limit", "pro", 5000, false},
		{"PRO plan - exceeds limit", "pro", 5001, true},
		{"ENTERPRISE plan - within limit", "enterprise", 8000, false},
		{"ENTERPRISE plan - at limit", "enterprise", 10000, false},
		{"ENTERPRISE plan - exceeds limit", "enterprise", 10001, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test user with the specified plan
			userID := "test-user-" + tt.userPlan
			userRepo.users[userID] = &model.User{
				ID:   userID,
				Name: "Test User",
				Plan: tt.userPlan,
			}

			// Create content with specified length (using multi-byte characters to test rune counting)
			content := strings.Repeat("あ", tt.contentLength) // Japanese character (3 bytes per character)

			req := connect.NewRequest(&postv1.CreatePostRequest{
				UserId:  userID,
				Title:   "Test Post",
				Content: content,
			})

			resp, err := handler.CreatePost(ctx, req)

			if tt.shouldFail {
				if err == nil {
					t.Errorf("Expected error for content length %d with plan %s, but got none", tt.contentLength, tt.userPlan)
				}
				if err != nil && connect.CodeOf(err) != connect.CodeInvalidArgument {
					t.Errorf("Expected CodeInvalidArgument, got %v", connect.CodeOf(err))
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for content length %d with plan %s: %v", tt.contentLength, tt.userPlan, err)
				}
				if resp != nil && resp.Msg.Post.XUserPlan == commonv1.UserPlan_USER_PLAN_UNSPECIFIED {
					t.Error("Expected XUserPlan to be set in response")
				}
			}
		})
	}
}

func TestPostHandler_CreatePost_UserNotFound(t *testing.T) {
	postRepo := newMockPostRepository()
	userRepo := newMockUserRepositoryForPost()
	handler := NewPostHandler(postRepo, userRepo)
	ctx := context.Background()

	req := connect.NewRequest(&postv1.CreatePostRequest{
		UserId:  "non-existent-user",
		Title:   "Test Post",
		Content: "Test Content",
	})

	_, err := handler.CreatePost(ctx, req)

	if err == nil {
		t.Fatal("Expected error when user not found")
	}

	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Errorf("Expected CodeNotFound, got %v", connect.CodeOf(err))
	}
}

func TestPostHandler_ListPosts_Pagination(t *testing.T) {
	postRepo := newMockPostRepository()
	userRepo := newMockUserRepositoryForPost()
	handler := NewPostHandler(postRepo, userRepo)
	ctx := context.Background()

	// Create test user
	userID := "test-user"
	userRepo.users[userID] = &model.User{
		ID:   userID,
		Name: "Test User",
		Plan: "free",
	}

	// Create some test posts
	for i := 0; i < 5; i++ {
		postRepo.posts["post-"+string(rune('0'+i))] = &model.Post{
			ID:      "post-" + string(rune('0'+i)),
			UserID:  userID,
			Title:   "Test Post",
			Content: "Content",
		}
	}

	req := connect.NewRequest(&postv1.ListPostsRequest{
		UserId:   userID,
		Page:     1,
		PageSize: 3,
	})

	resp, err := handler.ListPosts(ctx, req)
	if err != nil {
		t.Fatalf("ListPosts failed: %v", err)
	}

	if resp.Msg.Total != 5 {
		t.Errorf("Expected total 5, got %d", resp.Msg.Total)
	}

	if len(resp.Msg.Posts) != 3 {
		t.Errorf("Expected 3 posts in page, got %d", len(resp.Msg.Posts))
	}
}

func TestGetMaxContentLengthForPlan(t *testing.T) {
	tests := []struct {
		plan     string
		expected int
	}{
		{"free", 1000},
		{"pro", 5000},
		{"enterprise", 10000},
		{"unknown", 10000}, // default to enterprise
	}

	for _, tt := range tests {
		t.Run(tt.plan, func(t *testing.T) {
			result := getMaxContentLengthForPlan(tt.plan)
			if result != tt.expected {
				t.Errorf("getMaxContentLengthForPlan(%s) = %d, want %d", tt.plan, result, tt.expected)
			}
		})
	}
}

// Test to verify repository error handling
func TestPostHandler_CreatePost_RepositoryError(t *testing.T) {
	// Create a failing post repository
	failingRepo := &failingPostRepository{}
	userRepo := newMockUserRepositoryForPost()
	handler := NewPostHandler(failingRepo, userRepo)
	ctx := context.Background()

	// Create test user
	userID := "test-user"
	userRepo.users[userID] = &model.User{
		ID:   userID,
		Name: "Test User",
		Plan: "free",
	}

	req := connect.NewRequest(&postv1.CreatePostRequest{
		UserId:  userID,
		Title:   "Test Post",
		Content: "Content",
	})

	_, err := handler.CreatePost(ctx, req)

	if err == nil {
		t.Fatal("Expected error from repository")
	}

	if connect.CodeOf(err) != connect.CodeInternal {
		t.Errorf("Expected CodeInternal, got %v", connect.CodeOf(err))
	}
}

// Failing repository for error testing
type failingPostRepository struct{}

func (f *failingPostRepository) Create(ctx context.Context, post *model.Post) error {
	return errors.New("repository error")
}

func (f *failingPostRepository) List(ctx context.Context, userID string, page, pageSize int) ([]*model.Post, int, error) {
	return nil, 0, errors.New("repository error")
}

func (f *failingPostRepository) GetByID(ctx context.Context, id string) (*model.Post, error) {
	return nil, errors.New("repository error")
}
