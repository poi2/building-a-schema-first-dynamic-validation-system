package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/poi2/building-a-schema-first-dynamic-validation-system/services/be/internal/model"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	ctx := context.Background()
	// Use test database URL from environment or default
	dbURL := os.Getenv("CELO_DB_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5433/be?sslmode=disable"
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Skipf("Failed to connect to test database: %v", err)
	}

	// Ping to verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("Failed to ping test database: %v", err)
	}

	// Create table
	migration := `
		CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(255) NOT NULL,
			plan VARCHAR(20) NOT NULL CHECK (plan IN ('free', 'pro', 'enterprise')),
			created_at TIMESTAMP WITH TIME ZONE NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL
		);
	`
	if _, err := pool.Exec(ctx, migration); err != nil {
		pool.Close()
		t.Fatalf("Failed to create table: %v", err)
	}

	// Clean up existing data
	if _, err := pool.Exec(ctx, "DELETE FROM users"); err != nil {
		pool.Close()
		t.Fatalf("Failed to clean up users table: %v", err)
	}

	return pool
}

func TestUserRepository_Create(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	repo := NewUserRepository(pool)
	ctx := context.Background()

	now := time.Now()
	user := &model.User{
		ID:        "test-user-id",
		Name:      "Test User",
		Email:     "test@example.com",
		Plan:      "free",
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Verify user was created
	created, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}

	if created.ID != user.ID {
		t.Errorf("Expected ID %s, got %s", user.ID, created.ID)
	}
	if created.Name != user.Name {
		t.Errorf("Expected Name %s, got %s", user.Name, created.Name)
	}
	if created.Email != user.Email {
		t.Errorf("Expected Email %s, got %s", user.Email, created.Email)
	}
	if created.Plan != user.Plan {
		t.Errorf("Expected Plan %s, got %s", user.Plan, created.Plan)
	}
}

func TestUserRepository_List(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	repo := NewUserRepository(pool)
	ctx := context.Background()

	// Create multiple users
	now := time.Now()
	for i := 0; i < 5; i++ {
		user := &model.User{
			ID:        "test-user-" + string(rune('0'+i)),
			Name:      "User " + string(rune('A'+i)),
			Email:     "user" + string(rune('0'+i)) + "@example.com",
			Plan:      "free",
			CreatedAt: now.Add(time.Duration(i) * time.Minute),
			UpdatedAt: now.Add(time.Duration(i) * time.Minute),
		}
		if err := repo.Create(ctx, user); err != nil {
			t.Fatalf("Failed to create user %d: %v", i, err)
		}
	}

	// Test pagination
	users, total, err := repo.List(ctx, 1, 3)
	if err != nil {
		t.Fatalf("Failed to list users: %v", err)
	}

	if total != 5 {
		t.Errorf("Expected total 5, got %d", total)
	}
	if len(users) != 3 {
		t.Errorf("Expected 3 users, got %d", len(users))
	}

	// Test second page
	users, total, err = repo.List(ctx, 2, 3)
	if err != nil {
		t.Fatalf("Failed to list users page 2: %v", err)
	}

	if total != 5 {
		t.Errorf("Expected total 5, got %d", total)
	}
	if len(users) != 2 {
		t.Errorf("Expected 2 users on page 2, got %d", len(users))
	}
}

func TestUserRepository_GetByID(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	repo := NewUserRepository(pool)
	ctx := context.Background()

	now := time.Now()
	user := &model.User{
		ID:        "test-user-id",
		Name:      "Test User",
		Email:     "test@example.com",
		Plan:      "pro",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Get by ID
	fetched, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to get user by ID: %v", err)
	}

	if fetched.ID != user.ID {
		t.Errorf("Expected ID %s, got %s", user.ID, fetched.ID)
	}
	if fetched.Plan != "pro" {
		t.Errorf("Expected Plan pro, got %s", fetched.Plan)
	}

	// Test non-existent user
	_, err = repo.GetByID(ctx, "non-existent-id")
	if err == nil {
		t.Error("Expected error for non-existent user, got nil")
	}
}
