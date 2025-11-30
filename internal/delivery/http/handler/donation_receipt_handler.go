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

type DonationReceiptHandler struct {
	receiptUC *usecase.DonationReceiptUseCase
}

func NewDonationReceiptHandler(receiptUC *usecase.DonationReceiptUseCase) *DonationReceiptHandler {
	return &DonationReceiptHandler{receiptUC: receiptUC}
}

// Create godoc
// @Summary Create new donation receipt
// @Description Create a new donation receipt with items
// @Tags Donation Receipts
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateDonationReceiptRequest true "Create Donation Receipt Request Body"
// @Success 201 {object} dto.DonationReceiptResponseWrapper
// @Failure 400 {object} dto.ErrorResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Router /api/v1/donation-receipts [post]
func (h *DonationReceiptHandler) Create(c *gin.Context) {
	var req dto.CreateDonationReceiptRequest
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
	items := make([]usecase.CreateDonationReceiptItemInput, len(req.Items))
	for i, item := range req.Items {
		items[i] = usecase.CreateDonationReceiptItemInput{
			FundType:    item.FundType,
			ZakatType:   item.ZakatType,
			PersonCount: item.PersonCount,
			Amount:      item.Amount,
			RiceKG:      item.RiceKG,
			Notes:       item.Notes,
		}
	}

	receipt, err := h.receiptUC.Create(usecase.CreateDonationReceiptInput{
		MuzakkiID:       req.MuzakkiID,
		ReceiptNumber:   req.ReceiptNumber,
		ReceiptDate:     req.ReceiptDate,
		PaymentMethod:   req.PaymentMethod,
		Notes:           req.Notes,
		CreatedByUserID: userID.(string),
		Items:           items,
	})
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Success(c, http.StatusCreated, "Donation receipt created", gin.H{
		"id":             receipt.ID,
		"receipt_number": receipt.ReceiptNumber,
		"receipt_date":   receipt.ReceiptDate,
		"total_amount":   receipt.TotalAmount,
	})
}

// FindAll godoc
// @Summary Get all donation receipts
// @Description Get list of donation receipts with pagination and filters
// @Tags Donation Receipts
// @Security BearerAuth
// @Produce json
// @Param date_from query string false "Filter by date from (YYYY-MM-DD)"
// @Param date_to query string false "Filter by date to (YYYY-MM-DD)"
// @Param fund_type query string false "Filter by fund type: zakat, infaq, sadaqah"
// @Param zakat_type query string false "Filter by zakat type: fitrah, maal"
// @Param payment_method query string false "Filter by payment method"
// @Param muzakki_id query string false "Filter by muzakki ID"
// @Param q query string false "Search in muzakki name or notes"
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(10)
// @Success 200 {object} dto.DonationReceiptListResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Failure 500 {object} dto.ErrorResponseWrapper
// @Router /api/v1/donation-receipts [get]
func (h *DonationReceiptHandler) FindAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))

	receipts, total, err := h.receiptUC.FindAll(repository.DonationReceiptFilter{
		DateFrom:      c.Query("date_from"),
		DateTo:        c.Query("date_to"),
		FundType:      c.Query("fund_type"),
		ZakatType:     c.Query("zakat_type"),
		PaymentMethod: c.Query("payment_method"),
		MuzakkiID:     c.Query("muzakki_id"),
		Query:         c.Query("q"),
		Page:          page,
		PerPage:       perPage,
	})
	if err != nil {
		response.InternalServerError(c, err.Error(), nil)
		return
	}

	var data []dto.DonationReceiptListItemResponse
	for _, r := range receipts {
		data = append(data, dto.DonationReceiptListItemResponse{
			ID:              r.ID,
			ReceiptNumber:   r.ReceiptNumber,
			ReceiptDate:     r.ReceiptDate,
			MuzakkiID:       r.MuzakkiID,
			MuzakkiName:     r.Muzakki.Name,
			PaymentMethod:   r.PaymentMethod,
			TotalAmount:     r.TotalAmount,
			Notes:           r.Notes,
			CreatedByUserID: r.CreatedByUserID,
			CreatedAt:       r.CreatedAt,
			UpdatedAt:       r.UpdatedAt,
		})
	}

	response.Success(c, http.StatusOK, "Get all donation receipts successful", gin.H{
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
// @Summary Get donation receipt by ID
// @Description Get a single donation receipt with all items
// @Tags Donation Receipts
// @Security BearerAuth
// @Produce json
// @Param id path string true "Donation Receipt ID"
// @Success 200 {object} dto.DonationReceiptResponseWrapper
// @Failure 400 {object} dto.ErrorResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Router /api/v1/donation-receipts/{id} [get]
func (h *DonationReceiptHandler) FindByID(c *gin.Context) {
	id := c.Param("id")

	receipt, err := h.receiptUC.FindByID(id)
	if err != nil {
		response.BadRequest(c, "Donation receipt not found", nil)
		return
	}

	// Convert items
	items := make([]dto.DonationReceiptItemResponse, len(receipt.Items))
	for i, item := range receipt.Items {
		items[i] = dto.DonationReceiptItemResponse{
			ID:          item.ID,
			FundType:    item.FundType,
			ZakatType:   item.ZakatType,
			PersonCount: item.PersonCount,
			Amount:      item.Amount,
			RiceKG:      item.RiceKG,
			Notes:       item.Notes,
		}
	}

	response.Success(c, http.StatusOK, "Get donation receipt successful", dto.DonationReceiptResponse{
		ID:            receipt.ID,
		ReceiptNumber: receipt.ReceiptNumber,
		ReceiptDate:   receipt.ReceiptDate,
		Muzakki: dto.MuzakkiInfo{
			ID:       receipt.Muzakki.ID,
			FullName: receipt.Muzakki.Name,
		},
		PaymentMethod: receipt.PaymentMethod,
		TotalAmount:   receipt.TotalAmount,
		Notes:         receipt.Notes,
		CreatedByUser: dto.UserInfo{
			ID:       receipt.CreatedByUser.ID,
			FullName: receipt.CreatedByUser.Name,
		},
		Items:     items,
		CreatedAt: receipt.CreatedAt,
		UpdatedAt: receipt.UpdatedAt,
	})
}

// Update godoc
// @Summary Update donation receipt
// @Description Update an existing donation receipt
// @Tags Donation Receipts
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Donation Receipt ID"
// @Param request body dto.UpdateDonationReceiptRequest true "Update Donation Receipt Request Body"
// @Success 200 {object} dto.DonationReceiptResponseWrapper
// @Failure 400 {object} dto.ErrorResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Router /api/v1/donation-receipts/{id} [put]
func (h *DonationReceiptHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateDonationReceiptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, gin.H{"error": err.Error()})
		return
	}

	// Convert items
	items := make([]usecase.CreateDonationReceiptItemInput, len(req.Items))
	for i, item := range req.Items {
		items[i] = usecase.CreateDonationReceiptItemInput{
			FundType:    item.FundType,
			ZakatType:   item.ZakatType,
			PersonCount: item.PersonCount,
			Amount:      item.Amount,
			RiceKG:      item.RiceKG,
			Notes:       item.Notes,
		}
	}

	receipt, err := h.receiptUC.Update(usecase.UpdateDonationReceiptInput{
		ID:            id,
		MuzakkiID:     req.MuzakkiID,
		ReceiptNumber: req.ReceiptNumber,
		ReceiptDate:   req.ReceiptDate,
		PaymentMethod: req.PaymentMethod,
		Notes:         req.Notes,
		Items:         items,
	})
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Success(c, http.StatusOK, "Donation receipt updated successfully", gin.H{
		"id":             receipt.ID,
		"receipt_number": receipt.ReceiptNumber,
		"receipt_date":   receipt.ReceiptDate,
		"total_amount":   receipt.TotalAmount,
	})
}

// Delete godoc
// @Summary Delete donation receipt
// @Description Delete a donation receipt
// @Tags Donation Receipts
// @Security BearerAuth
// @Produce json
// @Param id path string true "Donation Receipt ID"
// @Success 200 {object} dto.ResponseSuccess
// @Failure 400 {object} dto.ErrorResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Router /api/v1/donation-receipts/{id} [delete]
func (h *DonationReceiptHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.receiptUC.Delete(id); err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Success(c, http.StatusOK, "Donation receipt deleted successfully", nil)
}
