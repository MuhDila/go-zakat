package handler

import (
	"net/http"

	"go-zakat-be/internal/delivery/http/dto"
	"go-zakat-be/internal/domain/repository"
	"go-zakat-be/internal/usecase"
	"go-zakat-be/pkg/response"

	"github.com/gin-gonic/gin"
)

type ReportHandler struct {
	reportUC *usecase.ReportUseCase
}

func NewReportHandler(reportUC *usecase.ReportUseCase) *ReportHandler {
	return &ReportHandler{reportUC: reportUC}
}

// GetIncomeSummary godoc
// @Summary Get income summary report
// @Description Get income summary grouped by period with breakdown by fund type
// @Tags Reports
// @Security BearerAuth
// @Produce json
// @Param date_from query string false "Filter by date from (YYYY-MM-DD)"
// @Param date_to query string false "Filter by date to (YYYY-MM-DD)"
// @Param group_by query string false "Group by: daily, monthly" default(monthly)
// @Success 200 {object} dto.ReportResponseWrapper
// @Failure 400 {object} dto.ErrorResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Router /api/v1/reports/income-summary [get]
func (h *ReportHandler) GetIncomeSummary(c *gin.Context) {
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")
	groupBy := c.DefaultQuery("group_by", "monthly")

	results, err := h.reportUC.GetIncomeSummary(dateFrom, dateTo, groupBy)
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	// Convert to DTO
	data := make([]dto.IncomeSummaryResponse, len(results))
	for i, r := range results {
		data[i] = dto.IncomeSummaryResponse{
			Period:      r.Period,
			ZakatFitrah: r.ZakatFitrah,
			ZakatMaal:   r.ZakatMaal,
			Infaq:       r.Infaq,
			Sadaqah:     r.Sadaqah,
			Total:       r.Total,
		}
	}

	response.Success(c, http.StatusOK, "Get income summary successful", data)
}

// GetDistributionSummary godoc
// @Summary Get distribution summary report
// @Description Get distribution summary grouped by asnaf or program
// @Tags Reports
// @Security BearerAuth
// @Produce json
// @Param date_from query string false "Filter by date from (YYYY-MM-DD)"
// @Param date_to query string false "Filter by date to (YYYY-MM-DD)"
// @Param group_by query string true "Group by: asnaf, program"
// @Param source_fund_type query string false "Filter by source fund type"
// @Success 200 {object} dto.ReportResponseWrapper
// @Failure 400 {object} dto.ErrorResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Router /api/v1/reports/distribution-summary [get]
func (h *ReportHandler) GetDistributionSummary(c *gin.Context) {
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")
	groupBy := c.Query("group_by")
	sourceFundType := c.Query("source_fund_type")

	if groupBy == "" {
		response.BadRequest(c, "group_by parameter is required (asnaf or program)", nil)
		return
	}

	results, err := h.reportUC.GetDistributionSummary(dateFrom, dateTo, groupBy, sourceFundType)
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	// Convert to DTO based on groupBy
	var data interface{}
	if groupBy == "asnaf" {
		asnafResults := results.([]repository.DistributionSummaryByAsnafResult)
		asnafData := make([]dto.DistributionSummaryByAsnafResponse, len(asnafResults))
		for i, r := range asnafResults {
			asnafData[i] = dto.DistributionSummaryByAsnafResponse{
				AsnafName:        r.AsnafName,
				BeneficiaryCount: r.BeneficiaryCount,
				TotalAmount:      r.TotalAmount,
			}
		}
		data = asnafData
	} else {
		programResults := results.([]repository.DistributionSummaryByProgramResult)
		programData := make([]dto.DistributionSummaryByProgramResponse, len(programResults))
		for i, r := range programResults {
			programData[i] = dto.DistributionSummaryByProgramResponse{
				ProgramName:      r.ProgramName,
				SourceFundType:   r.SourceFundType,
				BeneficiaryCount: r.BeneficiaryCount,
				TotalAmount:      r.TotalAmount,
			}
		}
		data = programData
	}

	response.Success(c, http.StatusOK, "Get distribution summary successful", data)
}

// GetFundBalance godoc
// @Summary Get fund balance report
// @Description Get fund balance showing total in, total out, and balance for each fund type
// @Tags Reports
// @Security BearerAuth
// @Produce json
// @Param date_from query string false "Filter by date from (YYYY-MM-DD)"
// @Param date_to query string false "Filter by date to (YYYY-MM-DD)"
// @Success 200 {object} dto.ReportResponseWrapper
// @Failure 400 {object} dto.ErrorResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Router /api/v1/reports/fund-balance [get]
func (h *ReportHandler) GetFundBalance(c *gin.Context) {
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")

	results, err := h.reportUC.GetFundBalance(dateFrom, dateTo)
	if err != nil {
		response.InternalServerError(c, err.Error(), nil)
		return
	}

	// Convert to DTO
	data := make([]dto.FundBalanceResponse, len(results))
	for i, r := range results {
		data[i] = dto.FundBalanceResponse{
			FundType: r.FundType,
			TotalIn:  r.TotalIn,
			TotalOut: r.TotalOut,
			Balance:  r.Balance,
		}
	}

	response.Success(c, http.StatusOK, "Get fund balance successful", data)
}

// GetMustahiqHistory godoc
// @Summary Get mustahiq history report
// @Description Get distribution history for a specific mustahiq
// @Tags Reports
// @Security BearerAuth
// @Produce json
// @Param mustahiq_id path string true "Mustahiq ID"
// @Success 200 {object} dto.ReportResponseWrapper
// @Failure 400 {object} dto.ErrorResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Router /api/v1/reports/mustahiq-history/{mustahiq_id} [get]
func (h *ReportHandler) GetMustahiqHistory(c *gin.Context) {
	mustahiqID := c.Param("mustahiq_id")

	result, err := h.reportUC.GetMustahiqHistory(mustahiqID)
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	// Convert to DTO
	history := make([]dto.MustahiqHistoryItemResponse, len(result.History))
	for i, h := range result.History {
		history[i] = dto.MustahiqHistoryItemResponse{
			DistributionDate: h.DistributionDate,
			ProgramName:      h.ProgramName,
			SourceFundType:   h.SourceFundType,
			Amount:           h.Amount,
		}
	}

	data := dto.MustahiqHistoryResponse{
		Mustahiq: dto.MustahiqHistoryMustahiqInfo{
			ID:        result.MustahiqID,
			FullName:  result.FullName,
			AsnafName: result.AsnafName,
			Address:   result.Address,
		},
		History:       history,
		TotalReceived: result.TotalReceived,
	}

	response.Success(c, http.StatusOK, "Get mustahiq history successful", data)
}
