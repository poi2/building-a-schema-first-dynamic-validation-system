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

func TestYAMLUserRepository_Create(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "test_users.yaml")

	repo, err := NewYAMLUserRepository(yamlPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()
	user := &model.User{
		ID:        "test-id-123",
		Name:      "Test User",
		Email:     "test@example.com",
		Plan:      "free",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create user
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Verify the file exists
	if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
		t.Fatalf("YAML file was not created")
	}

	// Retrieve the user
	retrieved, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve user: %v", err)
	}

	if retrieved.ID != user.ID || retrieved.Name != user.Name || retrieved.Email != user.Email {
		t.Errorf("Retrieved user does not match: got %+v, want %+v", retrieved, user)
	}
}

func TestYAMLUserRepository_List(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "test_users.yaml")

	repo, err := NewYAMLUserRepository(yamlPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()

	// Create multiple users
	users := []*model.User{
		{ID: "user-1", Name: "User 1", Email: "user1@example.com", Plan: "free", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "user-2", Name: "User 2", Email: "user2@example.com", Plan: "pro", CreatedAt: time.Now().Add(1 * time.Second), UpdatedAt: time.Now()},
		{ID: "user-3", Name: "User 3", Email: "user3@example.com", Plan: "enterprise", CreatedAt: time.Now().Add(2 * time.Second), UpdatedAt: time.Now()},
	}

	for _, user := range users {
		if err := repo.Create(ctx, user); err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	// Test pagination - first page
	page1, total, err := repo.List(ctx, 1, 2)
	if err != nil {
		t.Fatalf("Failed to list users: %v", err)
	}

	if total != 3 {
		t.Errorf("Expected total 3, got %d", total)
	}

	if len(page1) != 2 {
		t.Fatalf("Expected 2 users on page 1, got %d", len(page1))
	}

	// Verify DESC order (newest first)
	if page1[0].ID != "user-3" {
		t.Errorf("Expected first user to be user-3 (newest), got %s", page1[0].ID)
	}
	if page1[1].ID != "user-2" {
		t.Errorf("Expected second user to be user-2, got %s", page1[1].ID)
	}

	// Test pagination - second page
	page2, _, err := repo.List(ctx, 2, 2)
	if err != nil {
		t.Fatalf("Failed to list users page 2: %v", err)
	}

	if len(page2) != 1 {
		t.Fatalf("Expected 1 user on page 2, got %d", len(page2))
	}

	if page2[0].ID != "user-1" {
		t.Errorf("Expected user-1 on page 2, got %s", page2[0].ID)
	}
}

func TestYAMLUserRepository_GetByID_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "test_users.yaml")

	repo, err := NewYAMLUserRepository(yamlPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()

	// Try to get a non-existent user
	_, err = repo.GetByID(ctx, "non-existent-id")
	if err == nil {
		t.Fatal("Expected error for non-existent user, got nil")
	}

	// Verify the error wraps os.ErrNotExist
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Expected error to wrap os.ErrNotExist, got: %v", err)
	}
}

func TestYAMLUserRepository_AtomicWrite(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "test_users.yaml")

	repo, err := NewYAMLUserRepository(yamlPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()
	user := &model.User{
		ID:        "atomic-test",
		Name:      "Atomic Test User",
		Email:     "atomic@example.com",
		Plan:      "pro",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create user
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Verify no temp files are left behind
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}

	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == "" && entry.Name() != "test_users.yaml" {
			t.Errorf("Found unexpected file (possible temp file): %s", entry.Name())
		}
	}

	// Verify the main file exists and has correct permissions
	info, err := os.Stat(yamlPath)
	if err != nil {
		t.Fatalf("Failed to stat YAML file: %v", err)
	}

	expectedPerms := os.FileMode(0644)
	if info.Mode().Perm() != expectedPerms {
		t.Errorf("Expected file permissions %v, got %v", expectedPerms, info.Mode().Perm())
	}
}

func TestYAMLUserRepository_FileCreation(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "subdir", "test_users.yaml")

	// Repository should create the directory if it doesn't exist
	repo, err := NewYAMLUserRepository(yamlPath)
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
	users, total, err := repo.List(ctx, 1, 10)
	if err != nil {
		t.Fatalf("Failed to list users: %v", err)
	}

	if total != 0 || len(users) != 0 {
		t.Errorf("Expected empty user list, got total=%d, len=%d", total, len(users))
	}
}
