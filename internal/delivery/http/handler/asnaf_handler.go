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

type AsnafHandler struct {
	asnafUC *usecase.AsnafUseCase
}

func NewAsnafHandler(asnafUC *usecase.AsnafUseCase) *AsnafHandler {
	return &AsnafHandler{asnafUC: asnafUC}
}

// Create godoc
// @Summary Create new asnaf
// @Description Create a new asnaf record
// @Tags Asnaf
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateAsnafRequest true "Create Asnaf Request Body"
// @Success 201 {object} dto.AsnafResponseWrapper
// @Failure 400 {object} dto.ErrorResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Router /api/v1/asnaf [post]
func (h *AsnafHandler) Create(c *gin.Context) {
	var req dto.CreateAsnafRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, gin.H{"error": err.Error()})
		return
	}

	asnaf, err := h.asnafUC.Create(usecase.CreateAsnafInput{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Success(c, http.StatusCreated, "Asnaf created successfully", dto.AsnafResponse{
		ID:          asnaf.ID,
		Name:        asnaf.Name,
		Description: asnaf.Description,
		CreatedAt:   asnaf.CreatedAt,
		UpdatedAt:   asnaf.UpdatedAt,
	})
}

// FindAll godoc
// @Summary Get all asnaf
// @Description Get list of asnaf with pagination and search
// @Tags Asnaf
// @Security BearerAuth
// @Produce json
// @Param q query string false "Search by name"
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(10)
// @Success 200 {object} dto.AsnafListResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Failure 500 {object} dto.ErrorResponseWrapper
// @Router /api/v1/asnaf [get]
func (h *AsnafHandler) FindAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	query := c.Query("q")

	asnafs, total, err := h.asnafUC.FindAll(repository.AsnafFilter{
		Query:   query,
		Page:    page,
		PerPage: perPage,
	})
	if err != nil {
		response.InternalServerError(c, err.Error(), nil)
		return
	}

	var data []dto.AsnafResponse
	for _, a := range asnafs {
		data = append(data, dto.AsnafResponse{
			ID:          a.ID,
			Name:        a.Name,
			Description: a.Description,
			CreatedAt:   a.CreatedAt,
			UpdatedAt:   a.UpdatedAt,
		})
	}

	response.Success(c, http.StatusOK, "Get all asnaf successful", gin.H{
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
// @Summary Get asnaf by ID
// @Description Get a single asnaf record by ID
// @Tags Asnaf
// @Security BearerAuth
// @Produce json
// @Param id path string true "Asnaf ID"
// @Success 200 {object} dto.AsnafResponseWrapper
// @Failure 400 {object} dto.ErrorResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Router /api/v1/asnaf/{id} [get]
func (h *AsnafHandler) FindByID(c *gin.Context) {
	id := c.Param("id")

	asnaf, err := h.asnafUC.FindByID(id)
	if err != nil {
		response.BadRequest(c, "Asnaf not found", nil)
		return
	}

	response.Success(c, http.StatusOK, "Get asnaf successful", dto.AsnafResponse{
		ID:          asnaf.ID,
		Name:        asnaf.Name,
		Description: asnaf.Description,
		CreatedAt:   asnaf.CreatedAt,
		UpdatedAt:   asnaf.UpdatedAt,
	})
}

// Update godoc
// @Summary Update asnaf
// @Description Update an existing asnaf record
// @Tags Asnaf
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Asnaf ID"
// @Param request body dto.UpdateAsnafRequest true "Update Asnaf Request Body"
// @Success 200 {object} dto.AsnafResponseWrapper
// @Failure 400 {object} dto.ErrorResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Router /api/v1/asnaf/{id} [put]
func (h *AsnafHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateAsnafRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, gin.H{"error": err.Error()})
		return
	}

	asnaf, err := h.asnafUC.Update(usecase.UpdateAsnafInput{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Success(c, http.StatusOK, "Asnaf updated successfully", dto.AsnafResponse{
		ID:          asnaf.ID,
		Name:        asnaf.Name,
		Description: asnaf.Description,
		CreatedAt:   asnaf.CreatedAt,
		UpdatedAt:   asnaf.UpdatedAt,
	})
}

// Delete godoc
// @Summary Delete asnaf
// @Description Delete an asnaf record
// @Tags Asnaf
// @Security BearerAuth
// @Produce json
// @Param id path string true "Asnaf ID"
// @Success 200 {object} dto.ResponseSuccess
// @Failure 400 {object} dto.ErrorResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Router /api/v1/asnaf/{id} [delete]
func (h *AsnafHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.asnafUC.Delete(id); err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Success(c, http.StatusOK, "Asnaf deleted successfully", nil)
}
