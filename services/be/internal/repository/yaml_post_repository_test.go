package repository

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/poi2/building-a-schema-first-dynamic-validation-system/services/be/internal/model"
)

func TestYAMLPostRepository_Create(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "test_posts.yaml")

	repo, err := NewYAMLPostRepository(yamlPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()
	post := &model.Post{
		ID:        "test-post-123",
		UserID:    "user-123",
		Title:     "Test Post",
		Content:   "This is a test post content",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create post
	if err := repo.Create(ctx, post); err != nil {
		t.Fatalf("Failed to create post: %v", err)
	}

	// Verify the file exists
	if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
		t.Fatalf("YAML file was not created")
	}

	// Retrieve the post
	retrieved, err := repo.GetByID(ctx, post.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve post: %v", err)
	}

	if retrieved.ID != post.ID || retrieved.UserID != post.UserID || retrieved.Title != post.Title {
		t.Errorf("Retrieved post does not match: got %+v, want %+v", retrieved, post)
	}
}

func TestYAMLPostRepository_List(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "test_posts.yaml")

	repo, err := NewYAMLPostRepository(yamlPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()

	// Create posts for multiple users
	posts := []*model.Post{
		{ID: "post-1", UserID: "user-1", Title: "Post 1", Content: "Content 1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "post-2", UserID: "user-2", Title: "Post 2", Content: "Content 2", CreatedAt: time.Now().Add(1 * time.Second), UpdatedAt: time.Now()},
		{ID: "post-3", UserID: "user-1", Title: "Post 3", Content: "Content 3", CreatedAt: time.Now().Add(2 * time.Second), UpdatedAt: time.Now()},
		{ID: "post-4", UserID: "user-1", Title: "Post 4", Content: "Content 4", CreatedAt: time.Now().Add(3 * time.Second), UpdatedAt: time.Now()},
	}

	for _, post := range posts {
		if err := repo.Create(ctx, post); err != nil {
			t.Fatalf("Failed to create post: %v", err)
		}
	}

	// Test pagination for user-1 (has 3 posts)
	page1, total, err := repo.List(ctx, "user-1", 1, 2)
	if err != nil {
		t.Fatalf("Failed to list posts: %v", err)
	}

	if total != 3 {
		t.Errorf("Expected total 3 for user-1, got %d", total)
	}

	if len(page1) != 2 {
		t.Fatalf("Expected 2 posts on page 1, got %d", len(page1))
	}

	// Verify DESC order (newest first)
	if page1[0].ID != "post-4" {
		t.Errorf("Expected first post to be post-4 (newest), got %s", page1[0].ID)
	}
	if page1[1].ID != "post-3" {
		t.Errorf("Expected second post to be post-3, got %s", page1[1].ID)
	}

	// Test pagination - second page
	page2, _, err := repo.List(ctx, "user-1", 2, 2)
	if err != nil {
		t.Fatalf("Failed to list posts page 2: %v", err)
	}

	if len(page2) != 1 {
		t.Fatalf("Expected 1 post on page 2, got %d", len(page2))
	}

	if page2[0].ID != "post-1" {
		t.Errorf("Expected post-1 on page 2, got %s", page2[0].ID)
	}

	// Test filtering by user-2
	user2Posts, total2, err := repo.List(ctx, "user-2", 1, 10)
	if err != nil {
		t.Fatalf("Failed to list posts for user-2: %v", err)
	}

	if total2 != 1 {
		t.Errorf("Expected total 1 for user-2, got %d", total2)
	}

	if len(user2Posts) != 1 || user2Posts[0].ID != "post-2" {
		t.Errorf("Expected post-2 for user-2, got %v", user2Posts)
	}
}

func TestYAMLPostRepository_GetByID_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "test_posts.yaml")

	repo, err := NewYAMLPostRepository(yamlPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()

	// Try to get a non-existent post
	_, err = repo.GetByID(ctx, "non-existent-id")
	if err == nil {
		t.Fatal("Expected error for non-existent post, got nil")
	}

	// Verify the error wraps os.ErrNotExist
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Expected error to wrap os.ErrNotExist, got: %v", err)
	}
}

func TestYAMLPostRepository_FileCreation(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "subdir", "test_posts.yaml")

	// Repository should create the directory if it doesn't exist
	repo, err := NewYAMLPostRepository(yamlPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Verify the directory and file were created
	if _, err := os.Stat(filepath.Dir(yamlPath)); os.IsNotExist(err) {
		t.Fatal("Expected directory to be created")
	}

	if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
		t.Fatal("Expected YAML file to be created")
	}

	// Verify we can perform operations
	ctx := context.Background()
	posts, total, err := repo.List(ctx, "any-user", 1, 10)
	if err != nil {
		t.Fatalf("Failed to list posts: %v", err)
	}

	if total != 0 || len(posts) != 0 {
		t.Errorf("Expected empty post list, got total=%d, len=%d", total, len(posts))
	}
}
