package handler

import (
	"net/http"
	"strconv"

	"go-zakat/internal/delivery/http/dto"
	"go-zakat/internal/domain/repository"
	"go-zakat/internal/usecase"
	"go-zakat/pkg/response"

	"github.com/gin-gonic/gin"
)

type MustahiqHandler struct {
	mustahiqUC *usecase.MustahiqUseCase
}

func NewMustahiqHandler(mustahiqUC *usecase.MustahiqUseCase) *MustahiqHandler {
	return &MustahiqHandler{mustahiqUC: mustahiqUC}
}

// Create godoc
// @Summary Create new mustahiq
// @Description Create a new mustahiq record
// @Tags Mustahiq
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateMustahiqRequest true "Create Mustahiq Request Body"
// @Success 201 {object} dto.MustahiqResponseWrapper
// @Failure 400 {object} dto.ErrorResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Router /api/v1/mustahiq [post]
func (h *MustahiqHandler) Create(c *gin.Context) {
	var req dto.CreateMustahiqRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, gin.H{"error": err.Error()})
		return
	}

	mustahiq, err := h.mustahiqUC.Create(usecase.CreateMustahiqInput{
		Name:        req.Name,
		PhoneNumber: req.PhoneNumber,
		Address:     req.Address,
		AsnafID:     req.AsnafID,
		Status:      req.Status,
		Description: req.Description,
	})
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Success(c, http.StatusCreated, "Mustahiq created successfully", dto.MustahiqResponse{
		ID:          mustahiq.ID,
		Name:        mustahiq.Name,
		PhoneNumber: mustahiq.PhoneNumber,
		Address:     mustahiq.Address,
		Asnaf: dto.AsnafInfo{
			ID:   mustahiq.Asnaf.ID,
			Name: mustahiq.Asnaf.Name,
		},
		Status:      mustahiq.Status,
		Description: mustahiq.Description,
		CreatedAt:   mustahiq.CreatedAt,
		UpdatedAt:   mustahiq.UpdatedAt,
	})
}

// FindAll godoc
// @Summary Get all mustahiq
// @Description Get list of mustahiq with pagination, search, and status filter
// @Tags Mustahiq
// @Security BearerAuth
// @Produce json
// @Param q query string false "Search by name or address"
// @Param status query string false "Filter by status: active, inactive, pending"
// @Param asnafID query string false "Filter by asnaf ID"
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(10)
// @Success 200 {object} dto.MustahiqListResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Failure 500 {object} dto.ErrorResponseWrapper
// @Router /api/v1/mustahiq [get]
func (h *MustahiqHandler) FindAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	query := c.Query("q")
	status := c.Query("status")
	asnafID := c.Query("asnafID")

	mustahiqs, total, err := h.mustahiqUC.FindAll(repository.MustahiqFilter{
		Query:   query,
		Status:  status,
		AsnafID: asnafID,
		Page:    page,
		PerPage: perPage,
	})
	if err != nil {
		response.InternalServerError(c, err.Error(), nil)
		return
	}

	var data []dto.MustahiqResponse
	for _, m := range mustahiqs {
		data = append(data, dto.MustahiqResponse{
			ID:          m.ID,
			Name:        m.Name,
			PhoneNumber: m.PhoneNumber,
			Address:     m.Address,
			Asnaf: dto.AsnafInfo{
				ID:   m.Asnaf.ID,
				Name: m.Asnaf.Name,
			},
			Status:      m.Status,
			Description: m.Description,
			CreatedAt:   m.CreatedAt,
			UpdatedAt:   m.UpdatedAt,
		})
	}

	response.Success(c, http.StatusOK, "Get all mustahiq successful", gin.H{
		"data":       data,
		"total":      total,
		"page":       page,
		"per_page":   perPage,
		"total_page": (total + int64(perPage) - 1) / int64(perPage),
	})
}

// FindByID godoc
// @Summary Get mustahiq by ID
// @Description Get a single mustahiq record by ID
// @Tags Mustahiq
// @Security BearerAuth
// @Produce json
// @Param id path string true "Mustahiq ID"
// @Success 200 {object} dto.MustahiqResponseWrapper
// @Failure 400 {object} dto.ErrorResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Router /api/v1/mustahiq/{id} [get]
func (h *MustahiqHandler) FindByID(c *gin.Context) {
	id := c.Param("id")

	mustahiq, err := h.mustahiqUC.FindByID(id)
	if err != nil {
		response.BadRequest(c, "Mustahiq not found", nil)
		return
	}

	response.Success(c, http.StatusOK, "Get mustahiq successful", dto.MustahiqResponse{
		ID:          mustahiq.ID,
		Name:        mustahiq.Name,
		PhoneNumber: mustahiq.PhoneNumber,
		Address:     mustahiq.Address,
		Asnaf: dto.AsnafInfo{
			ID:   mustahiq.Asnaf.ID,
			Name: mustahiq.Asnaf.Name,
		},
		Status:      mustahiq.Status,
		Description: mustahiq.Description,
		CreatedAt:   mustahiq.CreatedAt,
		UpdatedAt:   mustahiq.UpdatedAt,
	})
}

// Update godoc
// @Summary Update mustahiq
// @Description Update an existing mustahiq record
// @Tags Mustahiq
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Mustahiq ID"
// @Param request body dto.UpdateMustahiqRequest true "Update Mustahiq Request Body"
// @Success 200 {object} dto.MustahiqResponseWrapper
// @Failure 400 {object} dto.ErrorResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Router /api/v1/mustahiq/{id} [put]
func (h *MustahiqHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateMustahiqRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, gin.H{"error": err.Error()})
		return
	}

	mustahiq, err := h.mustahiqUC.Update(usecase.UpdateMustahiqInput{
		ID:          id,
		Name:        req.Name,
		PhoneNumber: req.PhoneNumber,
		Address:     req.Address,
		AsnafID:     req.AsnafID,
		Status:      req.Status,
		Description: req.Description,
	})
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Success(c, http.StatusOK, "Mustahiq updated successfully", dto.MustahiqResponse{
		ID:          mustahiq.ID,
		Name:        mustahiq.Name,
		PhoneNumber: mustahiq.PhoneNumber,
		Address:     mustahiq.Address,
		Asnaf: dto.AsnafInfo{
			ID:   mustahiq.Asnaf.ID,
			Name: mustahiq.Asnaf.Name,
		},
		Status:      mustahiq.Status,
		Description: mustahiq.Description,
		CreatedAt:   mustahiq.CreatedAt,
		UpdatedAt:   mustahiq.UpdatedAt,
	})
}

// Delete godoc
// @Summary Delete mustahiq
// @Description Delete a mustahiq record
// @Tags Mustahiq
// @Security BearerAuth
// @Produce json
// @Param id path string true "Mustahiq ID"
// @Success 200 {object} dto.ResponseSuccess
// @Failure 400 {object} dto.ErrorResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Router /api/v1/mustahiq/{id} [delete]
func (h *MustahiqHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.mustahiqUC.Delete(id); err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Success(c, http.StatusOK, "Mustahiq deleted successfully", nil)
}
