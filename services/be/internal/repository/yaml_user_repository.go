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
	if _, err := os.Stat(r.filePath); os.IsNotExist(err) {
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

// writeFile writes the data to the YAML file
func (r *YAMLUserRepository) writeFile(data *yamlData) error {
	yamlBytes, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	if err := os.WriteFile(r.filePath, yamlBytes, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
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

	// Return slice of users (sorted by creation time - newest first)
	// Note: YAML preserves order, assuming users are added chronologically
	// For proper DESC order, we need to reverse
	reversedUsers := make([]*model.User, total)
	for i, user := range data.Users {
		reversedUsers[total-1-i] = user
	}

	return reversedUsers[offset:end], total, nil
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

	return nil, fmt.Errorf("user not found: %s", id)
}
