package handler

import (
	"net/http"
	"strconv"

	"go-zakat-be/internal/delivery/http/dto"
	"go-zakat-be/internal/usecase"
	"go-zakat-be/pkg/response"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userUC *usecase.UserUseCase
}

func NewUserHandler(userUC *usecase.UserUseCase) *UserHandler {
	return &UserHandler{userUC: userUC}
}

// FindAll godoc
// @Summary Get all users
// @Description Get all users with pagination and search (Admin only)
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Param q query string false "Search by name or email"
// @Param role query string false "Filter by role (admin, staf, viewer)"
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(10)
// @Success 200 {object} dto.UserListResponseWrapper
// @Failure 400 {object} dto.ErrorResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Failure 403 {object} dto.ErrorResponseWrapper
// @Router /api/v1/users [get]
func (h *UserHandler) FindAll(c *gin.Context) {
	query := c.Query("q")
	role := c.Query("role")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))

	users, total, err := h.userUC.FindAll(query, role, page, perPage)
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	// Convert to DTO
	userResponses := make([]dto.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = dto.UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			Role:      user.Role,
			GoogleID:  user.GoogleID,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	totalPage := (int(total) + perPage - 1) / perPage
	responseData := gin.H{
		"items": userResponses,
		"meta": dto.MetaResponse{
			Page:      page,
			PerPage:   perPage,
			Total:     int(total),
			TotalPage: totalPage,
		},
	}

	response.Success(c, http.StatusOK, "Get all users successful", responseData)
}

// FindByID godoc
// @Summary Get user by ID
// @Description Get user details by ID (Admin only)
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} dto.UserResponseWrapper
// @Failure 400 {object} dto.ErrorResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Failure 403 {object} dto.ErrorResponseWrapper
// @Failure 404 {object} dto.ErrorResponseWrapper
// @Router /api/v1/users/{id} [get]
func (h *UserHandler) FindByID(c *gin.Context) {
	userID := c.Param("id")

	user, err := h.userUC.FindByID(userID)
	if err != nil {
		response.BadRequest(c, "User not found", nil)
		return
	}

	userResponse := dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Role:      user.Role,
		GoogleID:  user.GoogleID,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(c, http.StatusOK, "Get user successful", userResponse)
}

// UpdateRole godoc
// @Summary Update user role
// @Description Update user role (Admin only)
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body dto.UpdateRoleRequest true "Role update request"
// @Success 200 {object} dto.UserResponseWrapper
// @Failure 400 {object} dto.ErrorResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Failure 403 {object} dto.ErrorResponseWrapper
// @Failure 404 {object} dto.ErrorResponseWrapper
// @Router /api/v1/users/{id}/role [put]
func (h *UserHandler) UpdateRole(c *gin.Context) {
	userID := c.Param("id")

	var req dto.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", nil)
		return
	}

	// Get current user ID from context
	currentUserID, _ := c.Get("user_id")

	user, err := h.userUC.UpdateRole(userID, req.Role, currentUserID.(string))
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	userResponse := dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Role:      user.Role,
		GoogleID:  user.GoogleID,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(c, http.StatusOK, "User role updated successfully. User needs to re-login to get new permissions.", userResponse)
}
