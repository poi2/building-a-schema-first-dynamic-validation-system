package handler

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	commonv1 "github.com/poi2/building-a-schema-first-dynamic-validation-system/pkg/gen/go/common/v1"
	userv1 "github.com/poi2/building-a-schema-first-dynamic-validation-system/pkg/gen/go/user/v1"
	"github.com/poi2/building-a-schema-first-dynamic-validation-system/services/be/internal/model"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// UserRepository interface for user data operations
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	List(ctx context.Context, page, pageSize int) ([]*model.User, int, error)
	GetByID(ctx context.Context, id string) (*model.User, error)
}

// UserHandler implements the UserService
type UserHandler struct {
	repo UserRepository
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(repo UserRepository) *UserHandler {
	return &UserHandler{repo: repo}
}

// CreateUser creates a new user
func (h *UserHandler) CreateUser(
	ctx context.Context,
	req *connect.Request[userv1.CreateUserRequest],
) (*connect.Response[userv1.CreateUserResponse], error) {
	// Generate UUID v7
	userID := uuid.NewString()

	// Get current time
	now := time.Now()

	// Convert proto enum to string
	planStr := userPlanToString(req.Msg.Plan)

	// Create user model
	user := &model.User{
		ID:        userID,
		Name:      req.Msg.Name,
		Email:     req.Msg.Email,
		Plan:      planStr,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Save to database
	if err := h.repo.Create(ctx, user); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert to proto response
	protoUser := &userv1.User{
		Id:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Plan:      req.Msg.Plan,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}

	return connect.NewResponse(&userv1.CreateUserResponse{
		User: protoUser,
	}), nil
}

// ListUsers lists users with pagination
func (h *UserHandler) ListUsers(
	ctx context.Context,
	req *connect.Request[userv1.ListUsersRequest],
) (*connect.Response[userv1.ListUsersResponse], error) {
	// Get pagination parameters (default values if not provided)
	page := req.Msg.Page
	if page == 0 {
		page = 1
	}
	pageSize := req.Msg.PageSize
	if pageSize == 0 {
		pageSize = 10
	}

	// Fetch users from repository
	users, total, err := h.repo.List(ctx, int(page), int(pageSize))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert to proto users
	protoUsers := make([]*userv1.User, 0, len(users))
	for _, user := range users {
		protoUsers = append(protoUsers, &userv1.User{
			Id:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			Plan:      stringToUserPlan(user.Plan),
			CreatedAt: timestamppb.New(user.CreatedAt),
			UpdatedAt: timestamppb.New(user.UpdatedAt),
		})
	}

	return connect.NewResponse(&userv1.ListUsersResponse{
		Users: protoUsers,
		Total: int32(total),
	}), nil
}

// Helper function to convert proto UserPlan enum to string
func userPlanToString(plan commonv1.UserPlan) string {
	switch plan {
	case commonv1.UserPlan_USER_PLAN_FREE:
		return "free"
	case commonv1.UserPlan_USER_PLAN_PRO:
		return "pro"
	case commonv1.UserPlan_USER_PLAN_ENTERPRISE:
		return "enterprise"
	default:
		return "free" // default to free plan
	}
}

// Helper function to convert string to proto UserPlan enum
func stringToUserPlan(plan string) commonv1.UserPlan {
	switch plan {
	case "free":
		return commonv1.UserPlan_USER_PLAN_FREE
	case "pro":
		return commonv1.UserPlan_USER_PLAN_PRO
	case "enterprise":
		return commonv1.UserPlan_USER_PLAN_ENTERPRISE
	default:
		return commonv1.UserPlan_USER_PLAN_FREE
	}
}
