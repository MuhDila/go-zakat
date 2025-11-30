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

type DistributionHandler struct {
	distributionUC *usecase.DistributionUseCase
}

func NewDistributionHandler(distributionUC *usecase.DistributionUseCase) *DistributionHandler {
	return &DistributionHandler{distributionUC: distributionUC}
}

// Create godoc
// @Summary Create new distribution
// @Description Create a new distribution with items
// @Tags Distributions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateDistributionRequest true "Create Distribution Request Body"
// @Success 201 {object} dto.DistributionResponseWrapper
// @Failure 400 {object} dto.ErrorResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Router /api/v1/distributions [post]
func (h *DistributionHandler) Create(c *gin.Context) {
	var req dto.CreateDistributionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated", nil)
		return
	}

	// Convert items
	items := make([]usecase.CreateDistributionItemInput, len(req.Items))
	for i, item := range req.Items {
		items[i] = usecase.CreateDistributionItemInput{
			MustahiqID: item.MustahiqID,
			Amount:     item.Amount,
			Notes:      item.Notes,
		}
	}

	distribution, err := h.distributionUC.Create(usecase.CreateDistributionInput{
		DistributionDate: req.DistributionDate,
		ProgramID:        req.ProgramID,
		SourceFundType:   req.SourceFundType,
		Notes:            req.Notes,
		CreatedByUserID:  userID.(string),
		Items:            items,
	})
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Success(c, http.StatusCreated, "Distribution created", gin.H{
		"id":                distribution.ID,
		"distribution_date": distribution.DistributionDate,
		"total_amount":      distribution.TotalAmount,
	})
}

// FindAll godoc
// @Summary Get all distributions
// @Description Get list of distributions with pagination and filters
// @Tags Distributions
// @Security BearerAuth
// @Produce json
// @Param date_from query string false "Filter by date from (YYYY-MM-DD)"
// @Param date_to query string false "Filter by date to (YYYY-MM-DD)"
// @Param source_fund_type query string false "Filter by source fund type: zakat_fitrah, zakat_maal, infaq, sadaqah"
// @Param program_id query string false "Filter by program ID"
// @Param q query string false "Search in program name or notes"
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(10)
// @Success 200 {object} dto.DistributionListResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Failure 500 {object} dto.ErrorResponseWrapper
// @Router /api/v1/distributions [get]
func (h *DistributionHandler) FindAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))

	distributions, total, err := h.distributionUC.FindAll(repository.DistributionFilter{
		DateFrom:       c.Query("date_from"),
		DateTo:         c.Query("date_to"),
		SourceFundType: c.Query("source_fund_type"),
		ProgramID:      c.Query("program_id"),
		Query:          c.Query("q"),
		Page:           page,
		PerPage:        perPage,
	})
	if err != nil {
		response.InternalServerError(c, err.Error(), nil)
		return
	}

	var data []dto.DistributionListItemResponse
	for _, d := range distributions {
		item := dto.DistributionListItemResponse{
			ID:               d.ID,
			DistributionDate: d.DistributionDate,
			SourceFundType:   d.SourceFundType,
			TotalAmount:      d.TotalAmount,
			BeneficiaryCount: int64(len(d.Items)), // Count from items loaded
			Notes:            d.Notes,
			CreatedAt:        d.CreatedAt,
			UpdatedAt:        d.UpdatedAt,
		}

		if d.Program != nil {
			item.ProgramID = &d.Program.ID
			item.ProgramName = d.Program.Name
		}

		data = append(data, item)
	}

	response.Success(c, http.StatusOK, "Get all distributions successful", gin.H{
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
// @Summary Get distribution by ID
// @Description Get a single distribution with all items
// @Tags Distributions
// @Security BearerAuth
// @Produce json
// @Param id path string true "Distribution ID"
// @Success 200 {object} dto.DistributionResponseWrapper
// @Failure 400 {object} dto.ErrorResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Router /api/v1/distributions/{id} [get]
func (h *DistributionHandler) FindByID(c *gin.Context) {
	id := c.Param("id")

	distribution, err := h.distributionUC.FindByID(id)
	if err != nil {
		response.BadRequest(c, "Distribution not found", nil)
		return
	}

	// Convert items
	items := make([]dto.DistributionItemResponse, len(distribution.Items))
	for i, item := range distribution.Items {
		items[i] = dto.DistributionItemResponse{
			ID:           item.ID,
			MustahiqID:   item.MustahiqID,
			MustahiqName: item.Mustahiq.Name,
			AsnafName:    item.Mustahiq.Asnaf.Name,
			Address:      item.Mustahiq.Address,
			Amount:       item.Amount,
			Notes:        item.Notes,
		}
	}

	resp := dto.DistributionResponse{
		ID:               distribution.ID,
		DistributionDate: distribution.DistributionDate,
		SourceFundType:   distribution.SourceFundType,
		TotalAmount:      distribution.TotalAmount,
		Notes:            distribution.Notes,
		CreatedByUser: dto.UserInfo{
			ID:       distribution.CreatedByUser.ID,
			FullName: distribution.CreatedByUser.Name,
		},
		Items:     items,
		CreatedAt: distribution.CreatedAt,
		UpdatedAt: distribution.UpdatedAt,
	}

	if distribution.Program != nil {
		resp.Program = &dto.ProgramInfo{
			ID:   distribution.Program.ID,
			Name: distribution.Program.Name,
		}
	}

	response.Success(c, http.StatusOK, "Get distribution successful", resp)
}

// Update godoc
// @Summary Update distribution
// @Description Update an existing distribution
// @Tags Distributions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Distribution ID"
// @Param request body dto.UpdateDistributionRequest true "Update Distribution Request Body"
// @Success 200 {object} dto.DistributionResponseWrapper
// @Failure 400 {object} dto.ErrorResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Router /api/v1/distributions/{id} [put]
func (h *DistributionHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateDistributionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, gin.H{"error": err.Error()})
		return
	}

	// Convert items
	items := make([]usecase.CreateDistributionItemInput, len(req.Items))
	for i, item := range req.Items {
		items[i] = usecase.CreateDistributionItemInput{
			MustahiqID: item.MustahiqID,
			Amount:     item.Amount,
			Notes:      item.Notes,
		}
	}

	distribution, err := h.distributionUC.Update(usecase.UpdateDistributionInput{
		ID:               id,
		DistributionDate: req.DistributionDate,
		ProgramID:        req.ProgramID,
		SourceFundType:   req.SourceFundType,
		Notes:            req.Notes,
		Items:            items,
	})
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Success(c, http.StatusOK, "Distribution updated successfully", gin.H{
		"id":                distribution.ID,
		"distribution_date": distribution.DistributionDate,
		"total_amount":      distribution.TotalAmount,
	})
}

// Delete godoc
// @Summary Delete distribution
// @Description Delete a distribution
// @Tags Distributions
// @Security BearerAuth
// @Produce json
// @Param id path string true "Distribution ID"
// @Success 200 {object} dto.ResponseSuccess
// @Failure 400 {object} dto.ErrorResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Router /api/v1/distributions/{id} [delete]
func (h *DistributionHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.distributionUC.Delete(id); err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Success(c, http.StatusOK, "Distribution deleted successfully", nil)
}
