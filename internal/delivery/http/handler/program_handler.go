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

type ProgramHandler struct {
	programUC *usecase.ProgramUseCase
}

func NewProgramHandler(programUC *usecase.ProgramUseCase) *ProgramHandler {
	return &ProgramHandler{programUC: programUC}
}

// Create godoc
// @Summary Create new program
// @Description Create a new program record
// @Tags Program
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateProgramRequest true "Create Program Request Body"
// @Success 201 {object} dto.ProgramResponseWrapper
// @Failure 400 {object} dto.ErrorResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Router /api/v1/programs [post]
func (h *ProgramHandler) Create(c *gin.Context) {
	var req dto.CreateProgramRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, gin.H{"error": err.Error()})
		return
	}

	program, err := h.programUC.Create(usecase.CreateProgramInput{
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		Active:      req.Active,
	})
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Success(c, http.StatusCreated, "Program created successfully", dto.ProgramResponse{
		ID:          program.ID,
		Name:        program.Name,
		Type:        program.Type,
		Description: program.Description,
		Active:      program.Active,
		CreatedAt:   program.CreatedAt,
		UpdatedAt:   program.UpdatedAt,
	})
}

// FindAll godoc
// @Summary Get all programs
// @Description Get list of programs with pagination, search, and filters
// @Tags Program
// @Security BearerAuth
// @Produce json
// @Param q query string false "Search by name"
// @Param type query string false "Filter by type"
// @Param active query boolean false "Filter by active status"
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(10)
// @Success 200 {object} dto.ProgramListResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Failure 500 {object} dto.ErrorResponseWrapper
// @Router /api/v1/programs [get]
func (h *ProgramHandler) FindAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	query := c.Query("q")
	programType := c.Query("type")

	var active *bool
	if activeStr := c.Query("active"); activeStr != "" {
		activeBool := activeStr == "true"
		active = &activeBool
	}

	programs, total, err := h.programUC.FindAll(repository.ProgramFilter{
		Query:   query,
		Type:    programType,
		Active:  active,
		Page:    page,
		PerPage: perPage,
	})
	if err != nil {
		response.InternalServerError(c, err.Error(), nil)
		return
	}

	var data []dto.ProgramResponse
	for _, p := range programs {
		data = append(data, dto.ProgramResponse{
			ID:          p.ID,
			Name:        p.Name,
			Type:        p.Type,
			Description: p.Description,
			Active:      p.Active,
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
		})
	}

	response.Success(c, http.StatusOK, "Get all programs successful", gin.H{
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
// @Summary Get program by ID
// @Description Get a single program record by ID
// @Tags Program
// @Security BearerAuth
// @Produce json
// @Param id path string true "Program ID"
// @Success 200 {object} dto.ProgramResponseWrapper
// @Failure 400 {object} dto.ErrorResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Router /api/v1/programs/{id} [get]
func (h *ProgramHandler) FindByID(c *gin.Context) {
	id := c.Param("id")

	program, err := h.programUC.FindByID(id)
	if err != nil {
		response.BadRequest(c, "Program not found", nil)
		return
	}

	response.Success(c, http.StatusOK, "Get program successful", dto.ProgramResponse{
		ID:          program.ID,
		Name:        program.Name,
		Type:        program.Type,
		Description: program.Description,
		Active:      program.Active,
		CreatedAt:   program.CreatedAt,
		UpdatedAt:   program.UpdatedAt,
	})
}

// Update godoc
// @Summary Update program
// @Description Update an existing program record
// @Tags Program
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Program ID"
// @Param request body dto.UpdateProgramRequest true "Update Program Request Body"
// @Success 200 {object} dto.ProgramResponseWrapper
// @Failure 400 {object} dto.ErrorResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Router /api/v1/programs/{id} [put]
func (h *ProgramHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateProgramRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, gin.H{"error": err.Error()})
		return
	}

	program, err := h.programUC.Update(usecase.UpdateProgramInput{
		ID:          id,
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		Active:      req.Active,
	})
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Success(c, http.StatusOK, "Program updated successfully", dto.ProgramResponse{
		ID:          program.ID,
		Name:        program.Name,
		Type:        program.Type,
		Description: program.Description,
		Active:      program.Active,
		CreatedAt:   program.CreatedAt,
		UpdatedAt:   program.UpdatedAt,
	})
}

// Delete godoc
// @Summary Delete program
// @Description Delete a program record
// @Tags Program
// @Security BearerAuth
// @Produce json
// @Param id path string true "Program ID"
// @Success 200 {object} dto.ResponseSuccess
// @Failure 400 {object} dto.ErrorResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Router /api/v1/programs/{id} [delete]
func (h *ProgramHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.programUC.Delete(id); err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Success(c, http.StatusOK, "Program deleted successfully", nil)
}
