package repository

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/poi2/building-a-schema-first-dynamic-validation-system/services/be/internal/model"
	"gopkg.in/yaml.v3"
)

// YAMLUserRepository handles user data persistence using YAML files
type YAMLUserRepository struct {
	filePath string
	mu       sync.RWMutex
}

// yamlData represents the structure of the YAML file
type yamlData struct {
	Users []*model.User `yaml:"users"`
}

// NewYAMLUserRepository creates a new YAMLUserRepository
func NewYAMLUserRepository(filePath string) (*YAMLUserRepository, error) {
	repo := &YAMLUserRepository{
		filePath: filePath,
	}

	// Initialize file if it doesn't exist
	if err := repo.initFile(); err != nil {
		return nil, fmt.Errorf("failed to initialize YAML file: %w", err)
	}

	return repo, nil
}

// initFile creates the YAML file with an empty users array if it doesn't exist
func (r *YAMLUserRepository) initFile() error {
	// Check if file exists
	if _, err := os.Stat(r.filePath); err != nil {
		if os.IsNotExist(err) {
			// Create directory if needed
			dir := filepath.Dir(r.filePath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}

			// Create empty YAML file
			data := yamlData{Users: []*model.User{}}
			if err := r.writeFile(&data); err != nil {
				return fmt.Errorf("failed to create initial file: %w", err)
			}

			return nil
		}

		return fmt.Errorf("failed to stat file: %w", err)
	}

	return nil
}

// readFile reads and parses the YAML file
func (r *YAMLUserRepository) readFile() (*yamlData, error) {
	file, err := os.ReadFile(r.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var data yamlData
	if err := yaml.Unmarshal(file, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	// Initialize users slice if nil
	if data.Users == nil {
		data.Users = []*model.User{}
	}

	return &data, nil
}

// writeFile writes the data to the YAML file atomically
func (r *YAMLUserRepository) writeFile(data *yamlData) error {
	yamlBytes, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	// Atomic write: write to a temp file in the same directory, then rename.
	dir := filepath.Dir(r.filePath)
	tmpFile, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	// Ensure the temp file is removed if something goes wrong before rename.
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(yamlBytes); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to sync temp file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	if err := os.Rename(tmpFile.Name(), r.filePath); err != nil {
		return fmt.Errorf("failed to replace YAML file: %w", err)
	}

	// Ensure file permissions are consistent with previous behavior.
	if err := os.Chmod(r.filePath, 0644); err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	return nil
}

// Create inserts a new user into the YAML file
func (r *YAMLUserRepository) Create(ctx context.Context, user *model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := r.readFile()
	if err != nil {
		return err
	}

	data.Users = append(data.Users, user)

	if err := r.writeFile(data); err != nil {
		return err
	}

	return nil
}

// List retrieves users with pagination
func (r *YAMLUserRepository) List(ctx context.Context, page, pageSize int) ([]*model.User, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data, err := r.readFile()
	if err != nil {
		return nil, 0, err
	}

	total := len(data.Users)

	// Calculate pagination
	offset := (page - 1) * pageSize
	if offset >= total {
		return []*model.User{}, total, nil
	}

	end := offset + pageSize
	if end > total {
		end = total
	}

	// For proper DESC order, we iterate from the end and collect only the requested window
	resultSize := end - offset
	pageUsers := make([]*model.User, 0, resultSize)

	// In the reversed view, index 0 corresponds to data.Users[total-1],
	// index 1 to data.Users[total-2], etc.
	start := total - 1 - offset    // first index in data.Users for this page
	stop := total - end            // inclusive lower bound index in data.Users
	for i := start; i >= stop; i-- { // walk backwards to maintain newest-first order
		pageUsers = append(pageUsers, data.Users[i])
	}

	return pageUsers, total, nil
}

// GetByID retrieves a user by ID
func (r *YAMLUserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data, err := r.readFile()
	if err != nil {
		return nil, err
	}

	for _, user := range data.Users {
		if user.ID == id {
			return user, nil
		}
	}

	return nil, fmt.Errorf("user %s: %w", id, os.ErrNotExist)
}
