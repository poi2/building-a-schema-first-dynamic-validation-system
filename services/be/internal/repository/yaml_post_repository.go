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

// YAMLPostRepository handles post data persistence using YAML files
type YAMLPostRepository struct {
	filePath string
	mu       sync.RWMutex
}

// postYAMLData represents the structure of the YAML file
type postYAMLData struct {
	Posts []*model.Post `yaml:"posts"`
}

// NewYAMLPostRepository creates a new YAMLPostRepository
func NewYAMLPostRepository(filePath string) (*YAMLPostRepository, error) {
	repo := &YAMLPostRepository{
		filePath: filePath,
	}

	// Initialize file if it doesn't exist
	if err := repo.initFile(); err != nil {
		return nil, fmt.Errorf("failed to initialize YAML file: %w", err)
	}

	return repo, nil
}

// initFile creates the YAML file with an empty posts array if it doesn't exist
func (r *YAMLPostRepository) initFile() error {
	// Check if file exists
	if _, err := os.Stat(r.filePath); err != nil {
		if os.IsNotExist(err) {
			// Create directory if needed
			dir := filepath.Dir(r.filePath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}

			// Create empty YAML file
			data := postYAMLData{Posts: []*model.Post{}}
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
func (r *YAMLPostRepository) readFile() (*postYAMLData, error) {
	file, err := os.ReadFile(r.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var data postYAMLData
	if err := yaml.Unmarshal(file, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	// Initialize posts slice if nil
	if data.Posts == nil {
		data.Posts = []*model.Post{}
	}

	return &data, nil
}

// writeFile writes the data to the YAML file atomically
func (r *YAMLPostRepository) writeFile(data *postYAMLData) error {
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

// Create inserts a new post into the YAML file
func (r *YAMLPostRepository) Create(ctx context.Context, post *model.Post) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := r.readFile()
	if err != nil {
		return err
	}

	data.Posts = append(data.Posts, post)

	if err := r.writeFile(data); err != nil {
		return err
	}

	return nil
}

// List retrieves posts for a specific user with pagination
func (r *YAMLPostRepository) List(ctx context.Context, userID string, page, pageSize int) ([]*model.Post, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data, err := r.readFile()
	if err != nil {
		return nil, 0, err
	}

	// Filter posts by user_id
	userPosts := make([]*model.Post, 0)
	for _, post := range data.Posts {
		if post.UserID == userID {
			userPosts = append(userPosts, post)
		}
	}

	total := len(userPosts)

	// Calculate pagination
	offset := (page - 1) * pageSize
	if offset >= total {
		return []*model.Post{}, total, nil
	}

	end := offset + pageSize
	if end > total {
		end = total
	}

	// For proper DESC order, we iterate from the end and collect only the requested window
	resultSize := end - offset
	pagePosts := make([]*model.Post, 0, resultSize)

	// In the reversed view, index 0 corresponds to userPosts[total-1],
	// index 1 to userPosts[total-2], etc.
	start := total - 1 - offset    // first index in userPosts for this page
	stop := total - end            // inclusive lower bound index in userPosts
	for i := start; i >= stop; i-- { // walk backwards to maintain newest-first order
		pagePosts = append(pagePosts, userPosts[i])
	}

	return pagePosts, total, nil
}

// GetByID retrieves a post by ID
func (r *YAMLPostRepository) GetByID(ctx context.Context, id string) (*model.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data, err := r.readFile()
	if err != nil {
		return nil, err
	}

	for _, post := range data.Posts {
		if post.ID == id {
			return post, nil
		}
	}

	return nil, fmt.Errorf("post %s: %w", id, os.ErrNotExist)
}
