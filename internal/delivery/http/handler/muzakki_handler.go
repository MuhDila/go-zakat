package handler

import (
	"net/http"
	"strconv"

	"go-zakat-be/internal/delivery/http/dto"
	"go-zakat-be/internal/domain/repository"
	"go-zakat-be/internal/usecase"
	"go-zakat-be/pkg/response"

	"github.com/gin-gonic/gin"
)

type MuzakkiHandler struct {
	muzakkiUC *usecase.MuzakkiUseCase
}

func NewMuzakkiHandler(muzakkiUC *usecase.MuzakkiUseCase) *MuzakkiHandler {
	return &MuzakkiHandler{muzakkiUC: muzakkiUC}
}

// Create godoc
// @Summary Create new muzakki
// @Description Create a new muzakki record
// @Tags Muzakki
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateMuzakkiRequest true "Create Muzakki Request Body"
// @Success 201 {object} dto.MuzakkiResponseWrapper
// @Failure 400 {object} dto.ErrorResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Router /api/v1/muzakki [post]
func (h *MuzakkiHandler) Create(c *gin.Context) {
	var req dto.CreateMuzakkiRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, gin.H{"error": err.Error()})
		return
	}

	muzakki, err := h.muzakkiUC.Create(usecase.CreateMuzakkiInput{
		Name:        req.Name,
		PhoneNumber: req.PhoneNumber,
		Address:     req.Address,
		Notes:       req.Notes,
	})
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Success(c, http.StatusCreated, "Muzakki created successfully", dto.MuzakkiResponse{
		ID:          muzakki.ID,
		Name:        muzakki.Name,
		PhoneNumber: muzakki.PhoneNumber,
		Address:     muzakki.Address,
		Notes:       muzakki.Notes,
		CreatedAt:   muzakki.CreatedAt,
		UpdatedAt:   muzakki.UpdatedAt,
	})
}

// FindAll godoc
// @Summary Get all muzakki
// @Description Get list of muzakki with pagination and search
// @Tags Muzakki
// @Security BearerAuth
// @Produce json
// @Param q query string false "Search by name or phone number"
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(10)
// @Success 200 {object} dto.MuzakkiListResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Failure 500 {object} dto.ErrorResponseWrapper
// @Router /api/v1/muzakki [get]
func (h *MuzakkiHandler) FindAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	query := c.Query("q")

	muzakkis, total, err := h.muzakkiUC.FindAll(repository.MuzakkiFilter{
		Query:   query,
		Page:    page,
		PerPage: perPage,
	})
	if err != nil {
		response.InternalServerError(c, err.Error(), nil)
		return
	}

	var data []dto.MuzakkiResponse
	for _, m := range muzakkis {
		data = append(data, dto.MuzakkiResponse{
			ID:          m.ID,
			Name:        m.Name,
			PhoneNumber: m.PhoneNumber,
			Address:     m.Address,
			Notes:       m.Notes,
			CreatedAt:   m.CreatedAt,
			UpdatedAt:   m.UpdatedAt,
		})
	}

	response.Success(c, http.StatusOK, "Get all muzakki successful", gin.H{
		"items": data,
		"meta": gin.H{
			"page":       page,
			"per_page":   perPage,
			"total":      total,
			"total_page": (total + int64(perPage) - 1) / int64(perPage),
		},
	})
}

// FindByID godoc
// @Summary Get muzakki by ID
// @Description Get a single muzakki record by ID
// @Tags Muzakki
// @Security BearerAuth
// @Produce json
// @Param id path string true "Muzakki ID"
// @Success 200 {object} dto.MuzakkiResponseWrapper
// @Failure 400 {object} dto.ErrorResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Router /api/v1/muzakki/{id} [get]
func (h *MuzakkiHandler) FindByID(c *gin.Context) {
	id := c.Param("id")

	muzakki, err := h.muzakkiUC.FindByID(id)
	if err != nil {
		response.BadRequest(c, "Muzakki not found", nil)
		return
	}

	response.Success(c, http.StatusOK, "Get muzakki successful", dto.MuzakkiResponse{
		ID:          muzakki.ID,
		Name:        muzakki.Name,
		PhoneNumber: muzakki.PhoneNumber,
		Address:     muzakki.Address,
		Notes:       muzakki.Notes,
		CreatedAt:   muzakki.CreatedAt,
		UpdatedAt:   muzakki.UpdatedAt,
	})
}

// Update godoc
// @Summary Update muzakki
// @Description Update an existing muzakki record
// @Tags Muzakki
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Muzakki ID"
// @Param request body dto.UpdateMuzakkiRequest true "Update Muzakki Request Body"
// @Success 200 {object} dto.MuzakkiResponseWrapper
// @Failure 400 {object} dto.ErrorResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Router /api/v1/muzakki/{id} [put]
func (h *MuzakkiHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateMuzakkiRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, gin.H{"error": err.Error()})
		return
	}

	muzakki, err := h.muzakkiUC.Update(usecase.UpdateMuzakkiInput{
		ID:          id,
		Name:        req.Name,
		PhoneNumber: req.PhoneNumber,
		Address:     req.Address,
		Notes:       req.Notes,
	})
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Success(c, http.StatusOK, "Muzakki updated successfully", dto.MuzakkiResponse{
		ID:          muzakki.ID,
		Name:        muzakki.Name,
		PhoneNumber: muzakki.PhoneNumber,
		Address:     muzakki.Address,
		Notes:       muzakki.Notes,
		CreatedAt:   muzakki.CreatedAt,
		UpdatedAt:   muzakki.UpdatedAt,
	})
}

// Delete godoc
// @Summary Delete muzakki
// @Description Delete a muzakki record
// @Tags Muzakki
// @Security BearerAuth
// @Produce json
// @Param id path string true "Muzakki ID"
// @Success 200 {object} dto.ResponseSuccess
// @Failure 400 {object} dto.ErrorResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Router /api/v1/muzakki/{id} [delete]
func (h *MuzakkiHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.muzakkiUC.Delete(id); err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Success(c, http.StatusOK, "Muzakki deleted successfully", nil)
}
