package handler

import (
	"context"
	"errors"
	"os"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	postv1 "github.com/poi2/building-a-schema-first-dynamic-validation-system/pkg/gen/go/post/v1"
	"github.com/poi2/building-a-schema-first-dynamic-validation-system/services/be/internal/model"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// PostRepository interface for post data operations
type PostRepository interface {
	Create(ctx context.Context, post *model.Post) error
	List(ctx context.Context, userID string, page, pageSize int) ([]*model.Post, int, error)
	GetByID(ctx context.Context, id string) (*model.Post, error)
}

// PostHandler implements the PostService
type PostHandler struct {
	postRepo PostRepository
	userRepo UserRepository
}

// NewPostHandler creates a new PostHandler
func NewPostHandler(postRepo PostRepository, userRepo UserRepository) *PostHandler {
	return &PostHandler{
		postRepo: postRepo,
		userRepo: userRepo,
	}
}

// CreatePost creates a new post with Context Enrichment
func (h *PostHandler) CreatePost(
	ctx context.Context,
	req *connect.Request[postv1.CreatePostRequest],
) (*connect.Response[postv1.CreatePostResponse], error) {
	// Context Enrichment: Get user's plan from user repository
	user, err := h.userRepo.GetByID(ctx, req.Msg.UserId)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Inject user plan into request for validation
	req.Msg.XUserPlan = stringToUserPlan(user.Plan)

	// Manual validation: Check content length based on user plan
	// Since interceptors run before the handler, we need to validate here
	maxContentLength := getMaxContentLengthForPlan(user.Plan)
	contentLength := len(req.Msg.Content)
	if contentLength > maxContentLength {
		return nil, connect.NewError(
			connect.CodeInvalidArgument,
			errors.New("content exceeds plan limit (FREE: 1000, PRO: 5000, ENTERPRISE: 10000 chars)"),
		)
	}

	// Generate UUID v7
	id, err := uuid.NewV7()
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	postID := id.String()

	// Get current time
	now := time.Now()

	// Create post model
	post := &model.Post{
		ID:        postID,
		UserID:    req.Msg.UserId,
		Title:     req.Msg.Title,
		Content:   req.Msg.Content,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Save to repository
	if err := h.postRepo.Create(ctx, post); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert to proto response
	protoPost := &postv1.Post{
		Id:        post.ID,
		UserId:    post.UserID,
		Title:     post.Title,
		Content:   post.Content,
		CreatedAt: timestamppb.New(post.CreatedAt),
		UpdatedAt: timestamppb.New(post.UpdatedAt),
		XUserPlan: stringToUserPlan(user.Plan),
	}

	return connect.NewResponse(&postv1.CreatePostResponse{
		Post: protoPost,
	}), nil
}

// ListPosts lists posts for a specific user with pagination
func (h *PostHandler) ListPosts(
	ctx context.Context,
	req *connect.Request[postv1.ListPostsRequest],
) (*connect.Response[postv1.ListPostsResponse], error) {
	// Get pagination parameters (validated by interceptor/proto)
	userID := req.Msg.UserId
	page := req.Msg.Page
	pageSize := req.Msg.PageSize

	// Fetch posts from repository
	posts, total, err := h.postRepo.List(ctx, userID, int(page), int(pageSize))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert to proto posts
	protoPosts := make([]*postv1.Post, 0, len(posts))
	for _, post := range posts {
		protoPosts = append(protoPosts, &postv1.Post{
			Id:        post.ID,
			UserId:    post.UserID,
			Title:     post.Title,
			Content:   post.Content,
			CreatedAt: timestamppb.New(post.CreatedAt),
			UpdatedAt: timestamppb.New(post.UpdatedAt),
		})
	}

	return connect.NewResponse(&postv1.ListPostsResponse{
		Posts: protoPosts,
		Total: int32(total),
	}), nil
}

// getMaxContentLengthForPlan returns the maximum content length based on user plan
func getMaxContentLengthForPlan(plan string) int {
	switch plan {
	case "free":
		return 1000
	case "pro":
		return 5000
	case "enterprise":
		return 10000
	default:
		return 10000 // default to enterprise limit
	}
}
