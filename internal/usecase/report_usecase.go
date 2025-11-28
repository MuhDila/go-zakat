package usecase

import (
	"errors"
	"go-zakat/internal/domain/repository"

	"github.com/go-playground/validator/v10"
)

type ReportUseCase struct {
	reportRepo repository.ReportRepository
	validator  *validator.Validate
}

func NewReportUseCase(reportRepo repository.ReportRepository, validator *validator.Validate) *ReportUseCase {
	return &ReportUseCase{
		reportRepo: reportRepo,
		validator:  validator,
	}
}

func (uc *ReportUseCase) GetIncomeSummary(dateFrom, dateTo, groupBy string) ([]repository.IncomeSummaryResult, error) {
	// Validate groupBy
	if groupBy != "" && groupBy != "daily" && groupBy != "monthly" {
		return nil, errors.New("group_by must be 'daily' or 'monthly'")
	}

	// Default to monthly
	if groupBy == "" {
		groupBy = "monthly"
	}

	return uc.reportRepo.GetIncomeSummary(dateFrom, dateTo, groupBy)
}

func (uc *ReportUseCase) GetDistributionSummary(dateFrom, dateTo, groupBy, sourceFundType string) (interface{}, error) {
	// Validate groupBy
	if groupBy != "asnaf" && groupBy != "program" {
		return nil, errors.New("group_by must be 'asnaf' or 'program'")
	}

	// Validate sourceFundType if provided
	if sourceFundType != "" {
		validTypes := []string{"zakat_fitrah", "zakat_maal", "infaq", "sadaqah"}
		valid := false
		for _, t := range validTypes {
			if sourceFundType == t {
				valid = true
				break
			}
		}
		if !valid {
			return nil, errors.New("source_fund_type must be one of: zakat_fitrah, zakat_maal, infaq, sadaqah")
		}
	}

	return uc.reportRepo.GetDistributionSummary(dateFrom, dateTo, groupBy, sourceFundType)
}

func (uc *ReportUseCase) GetFundBalance(dateFrom, dateTo string) ([]repository.FundBalanceResult, error) {
	return uc.reportRepo.GetFundBalance(dateFrom, dateTo)
}

func (uc *ReportUseCase) GetMustahiqHistory(mustahiqID string) (*repository.MustahiqHistoryResult, error) {
	if mustahiqID == "" {
		return nil, errors.New("mustahiq_id is required")
	}

	return uc.reportRepo.GetMustahiqHistory(mustahiqID)
}
