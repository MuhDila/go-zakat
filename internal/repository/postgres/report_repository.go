package postgres

import (
	"context"
	"errors"
	"time"

	"go-zakat-be/internal/domain/repository"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type ReportRepository struct {
	db  *pgxpool.Pool
	log *logrus.Logger
}

func NewReportRepository(db *pgxpool.Pool, log *logrus.Logger) *ReportRepository {
	return &ReportRepository{db: db, log: log}
}

func (r *ReportRepository) GetIncomeSummary(dateFrom, dateTo, groupBy string) ([]repository.IncomeSummaryResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var periodFormat string
	if groupBy == "daily" {
		periodFormat = "dr.receipt_date::TEXT"
	} else { // monthly (default)
		periodFormat = "TO_CHAR(dr.receipt_date, 'YYYY-MM')"
	}

	// Complex query with CASE WHEN to pivot fund_types into columns
	query := `
		SELECT 
			` + periodFormat + ` as period,
			COALESCE(SUM(CASE 
				WHEN dri.fund_type = 'zakat' AND dri.zakat_type = 'fitrah' THEN dri.amount 
				ELSE 0 
			END), 0) as zakat_fitrah,
			COALESCE(SUM(CASE 
				WHEN dri.fund_type = 'zakat' AND dri.zakat_type = 'maal' THEN dri.amount 
				ELSE 0 
			END), 0) as zakat_maal,
			COALESCE(SUM(CASE 
				WHEN dri.fund_type = 'infaq' THEN dri.amount 
				ELSE 0 
			END), 0) as infaq,
			COALESCE(SUM(CASE 
				WHEN dri.fund_type = 'sadaqah' THEN dri.amount 
				ELSE 0 
			END), 0) as sadaqah,
			COALESCE(SUM(dri.amount), 0) as total
		FROM donation_receipts dr
		INNER JOIN donation_receipt_items dri ON dr.id = dri.receipt_id
		WHERE 1=1
	`

	var args []interface{}
	argIdx := 1

	if dateFrom != "" {
		query += ` AND dr.receipt_date >= $` + string(rune(argIdx+'0'))
		args = append(args, dateFrom)
		argIdx++
	}
	if dateTo != "" {
		query += ` AND dr.receipt_date <= $` + string(rune(argIdx+'0'))
		args = append(args, dateTo)
		argIdx++
	}

	query += ` GROUP BY period ORDER BY period ASC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []repository.IncomeSummaryResult
	for rows.Next() {
		var result repository.IncomeSummaryResult
		err := rows.Scan(
			&result.Period, &result.ZakatFitrah, &result.ZakatMaal,
			&result.Infaq, &result.Sadaqah, &result.Total,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

func (r *ReportRepository) GetDistributionSummary(dateFrom, dateTo, groupBy, sourceFundType string) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	if groupBy == "asnaf" {
		return r.getDistributionSummaryByAsnaf(ctx, dateFrom, dateTo, sourceFundType)
	} else if groupBy == "program" {
		return r.getDistributionSummaryByProgram(ctx, dateFrom, dateTo, sourceFundType)
	}

	return nil, errors.New("invalid group_by parameter, must be 'asnaf' or 'program'")
}

func (r *ReportRepository) getDistributionSummaryByAsnaf(ctx context.Context, dateFrom, dateTo, sourceFundType string) ([]repository.DistributionSummaryByAsnafResult, error) {
	query := `
		SELECT 
			a.name as asnaf_name,
			COUNT(DISTINCT di.mustahiq_id) as beneficiary_count,
			COALESCE(SUM(di.amount), 0) as total_amount
		FROM distribution_items di
		INNER JOIN distributions d ON di.distribution_id = d.id
		INNER JOIN mustahiq m ON di.mustahiq_id = m.id
		INNER JOIN asnaf a ON m.asnafID = a.id
		WHERE 1=1
	`

	var args []interface{}
	argIdx := 1

	if dateFrom != "" {
		query += ` AND d.distribution_date >= $` + string(rune(argIdx+'0'))
		args = append(args, dateFrom)
		argIdx++
	}
	if dateTo != "" {
		query += ` AND d.distribution_date <= $` + string(rune(argIdx+'0'))
		args = append(args, dateTo)
		argIdx++
	}
	if sourceFundType != "" {
		query += ` AND d.source_fund_type = $` + string(rune(argIdx+'0'))
		args = append(args, sourceFundType)
		argIdx++
	}

	query += ` GROUP BY a.name ORDER BY total_amount DESC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []repository.DistributionSummaryByAsnafResult
	for rows.Next() {
		var result repository.DistributionSummaryByAsnafResult
		err := rows.Scan(&result.AsnafName, &result.BeneficiaryCount, &result.TotalAmount)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

func (r *ReportRepository) getDistributionSummaryByProgram(ctx context.Context, dateFrom, dateTo, sourceFundType string) ([]repository.DistributionSummaryByProgramResult, error) {
	query := `
		SELECT 
			COALESCE(p.name, 'No Program') as program_name,
			d.source_fund_type,
			COUNT(DISTINCT di.mustahiq_id) as beneficiary_count,
			COALESCE(SUM(di.amount), 0) as total_amount
		FROM distributions d
		LEFT JOIN programs p ON d.program_id = p.id
		INNER JOIN distribution_items di ON d.id = di.distribution_id
		WHERE 1=1
	`

	var args []interface{}
	argIdx := 1

	if dateFrom != "" {
		query += ` AND d.distribution_date >= $` + string(rune(argIdx+'0'))
		args = append(args, dateFrom)
		argIdx++
	}
	if dateTo != "" {
		query += ` AND d.distribution_date <= $` + string(rune(argIdx+'0'))
		args = append(args, dateTo)
		argIdx++
	}
	if sourceFundType != "" {
		query += ` AND d.source_fund_type = $` + string(rune(argIdx+'0'))
		args = append(args, sourceFundType)
		argIdx++
	}

	query += ` GROUP BY p.name, d.source_fund_type ORDER BY total_amount DESC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []repository.DistributionSummaryByProgramResult
	for rows.Next() {
		var result repository.DistributionSummaryByProgramResult
		err := rows.Scan(&result.ProgramName, &result.SourceFundType, &result.BeneficiaryCount, &result.TotalAmount)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

func (r *ReportRepository) GetFundBalance(dateFrom, dateTo string) ([]repository.FundBalanceResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	// Query to get total IN and OUT for each fund type
	query := `
		WITH income AS (
			SELECT 
				CASE 
					WHEN dri.fund_type = 'zakat' AND dri.zakat_type = 'fitrah' THEN 'zakat_fitrah'
					WHEN dri.fund_type = 'zakat' AND dri.zakat_type = 'maal' THEN 'zakat_maal'
					WHEN dri.fund_type = 'infaq' THEN 'infaq'
					WHEN dri.fund_type = 'sadaqah' THEN 'sadaqah'
				END as fund_type,
				COALESCE(SUM(dri.amount), 0) as total_in
			FROM donation_receipts dr
			INNER JOIN donation_receipt_items dri ON dr.id = dri.receipt_id
			WHERE 1=1
	`

	var args []interface{}
	argIdx := 1

	if dateFrom != "" {
		query += ` AND dr.receipt_date >= $` + string(rune(argIdx+'0'))
		args = append(args, dateFrom)
		argIdx++
	}
	if dateTo != "" {
		query += ` AND dr.receipt_date <= $` + string(rune(argIdx+'0'))
		args = append(args, dateTo)
		argIdx++
	}

	query += `
			GROUP BY dri.fund_type, dri.zakat_type
		),
		outgoing AS (
			SELECT 
				d.source_fund_type as fund_type,
				COALESCE(SUM(d.total_amount), 0) as total_out
			FROM distributions d
			WHERE 1=1
	`

	// Reset argIdx for outgoing query (same date params)
	outgoingArgIdx := 1
	if dateFrom != "" {
		query += ` AND d.distribution_date >= $` + string(rune(outgoingArgIdx+'0'))
		outgoingArgIdx++
	}
	if dateTo != "" {
		query += ` AND d.distribution_date <= $` + string(rune(outgoingArgIdx+'0'))
		outgoingArgIdx++
	}

	query += `
			GROUP BY d.source_fund_type
		),
		all_fund_types AS (
			SELECT 'zakat_fitrah' as fund_type
			UNION SELECT 'zakat_maal'
			UNION SELECT 'infaq'
			UNION SELECT 'sadaqah'
		)
		SELECT 
			aft.fund_type,
			COALESCE(i.total_in, 0) as total_in,
			COALESCE(o.total_out, 0) as total_out,
			COALESCE(i.total_in, 0) - COALESCE(o.total_out, 0) as balance
		FROM all_fund_types aft
		LEFT JOIN income i ON aft.fund_type = i.fund_type
		LEFT JOIN outgoing o ON aft.fund_type = o.fund_type
		ORDER BY aft.fund_type
	`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []repository.FundBalanceResult
	for rows.Next() {
		var result repository.FundBalanceResult
		err := rows.Scan(&result.FundType, &result.TotalIn, &result.TotalOut, &result.Balance)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

func (r *ReportRepository) GetMustahiqHistory(mustahiqID string) (*repository.MustahiqHistoryResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	// Get mustahiq info
	mustahiqQuery := `
		SELECT m.id, m.name, a.name, m.address
		FROM mustahiq m
		INNER JOIN asnaf a ON m.asnafID = a.id
		WHERE m.id = $1
		LIMIT 1
	`

	result := &repository.MustahiqHistoryResult{}
	err := r.db.QueryRow(ctx, mustahiqQuery, mustahiqID).Scan(
		&result.MustahiqID, &result.FullName, &result.AsnafName, &result.Address,
	)
	if err != nil {
		return nil, errors.New("mustahiq not found")
	}

	// Get distribution history
	historyQuery := `
		SELECT 
			d.distribution_date,
			COALESCE(p.name, 'No Program') as program_name,
			d.source_fund_type,
			di.amount
		FROM distribution_items di
		INNER JOIN distributions d ON di.distribution_id = d.id
		LEFT JOIN programs p ON d.program_id = p.id
		WHERE di.mustahiq_id = $1
		ORDER BY d.distribution_date DESC
	`

	rows, err := r.db.Query(ctx, historyQuery, mustahiqID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []repository.MustahiqHistoryItem
	var totalReceived float64

	for rows.Next() {
		var item repository.MustahiqHistoryItem
		var distributionDate time.Time
		err := rows.Scan(&distributionDate, &item.ProgramName, &item.SourceFundType, &item.Amount)
		if err != nil {
			return nil, err
		}
		// Convert time.Time to YYYY-MM-DD string
		item.DistributionDate = distributionDate.Format("2006-01-02")
		history = append(history, item)
		totalReceived += item.Amount
	}

	result.History = history
	result.TotalReceived = totalReceived

	return result, nil
}
