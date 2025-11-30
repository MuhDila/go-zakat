package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go-zakat-be/internal/domain/entity"
	"go-zakat-be/internal/domain/repository"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type DistributionRepository struct {
	db  *pgxpool.Pool
	log *logrus.Logger
}

func NewDistributionRepository(db *pgxpool.Pool, log *logrus.Logger) *DistributionRepository {
	return &DistributionRepository{db: db, log: log}
}

func (r *DistributionRepository) FindAll(filter repository.DistributionFilter) ([]*entity.Distribution, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	// Base query with JOINs and beneficiary count subquery
	query := `
		SELECT d.id, d.distribution_date, d.program_id, COALESCE(p.name, '') as program_name,
		       d.source_fund_type, d.total_amount, d.notes,
		       (SELECT COUNT(*) FROM distribution_items WHERE distribution_id = d.id) as beneficiary_count,
		       d.created_at, d.updated_at
		FROM distributions d
		LEFT JOIN programs p ON d.program_id = p.id
	`

	countQuery := `
		SELECT COUNT(*)
		FROM distributions d
		LEFT JOIN programs p ON d.program_id = p.id
	`

	var args []interface{}
	argIdx := 1
	var conditions []string

	// Filter by date range
	if filter.DateFrom != "" {
		conditions = append(conditions, fmt.Sprintf("d.distribution_date >= $%d", argIdx))
		args = append(args, filter.DateFrom)
		argIdx++
	}
	if filter.DateTo != "" {
		conditions = append(conditions, fmt.Sprintf("d.distribution_date <= $%d", argIdx))
		args = append(args, filter.DateTo)
		argIdx++
	}

	// Filter by source_fund_type
	if filter.SourceFundType != "" {
		conditions = append(conditions, fmt.Sprintf("d.source_fund_type = $%d", argIdx))
		args = append(args, filter.SourceFundType)
		argIdx++
	}

	// Filter by program_id
	if filter.ProgramID != "" {
		conditions = append(conditions, fmt.Sprintf("d.program_id = $%d", argIdx))
		args = append(args, filter.ProgramID)
		argIdx++
	}

	// Search in program name or notes
	if filter.Query != "" {
		search := fmt.Sprintf("%%%s%%", filter.Query)
		conditions = append(conditions, fmt.Sprintf("(p.name ILIKE $%d OR d.notes ILIKE $%d)", argIdx, argIdx+1))
		args = append(args, search, search)
		argIdx += 2
	}

	// Add WHERE clause
	if len(conditions) > 0 {
		whereClause := " WHERE " + strings.Join(conditions, " AND ")
		query += whereClause
		countQuery += whereClause
	}

	// Get total count
	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Add ORDER BY and pagination
	query += " ORDER BY d.distribution_date DESC, d.created_at DESC"
	if filter.PerPage > 0 {
		offset := (filter.Page - 1) * filter.PerPage
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
		args = append(args, filter.PerPage, offset)
	}

	// Execute main query
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var distributions []*entity.Distribution
	for rows.Next() {
		d := &entity.Distribution{}
		var programName string
		var beneficiaryCount int64
		var distributionDate time.Time

		err := rows.Scan(
			&d.ID, &distributionDate, &d.ProgramID, &programName,
			&d.SourceFundType, &d.TotalAmount, &d.Notes, &beneficiaryCount,
			&d.CreatedAt, &d.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		// Convert time.Time to YYYY-MM-DD string
		d.DistributionDate = distributionDate.Format("2006-01-02")

		// Set program if exists
		if d.ProgramID != nil && *d.ProgramID != "" {
			d.Program = &entity.Program{
				ID:   *d.ProgramID,
				Name: programName,
			}
		}

		distributions = append(distributions, d)
	}

	return distributions, total, nil
}

func (r *DistributionRepository) FindByID(id string) (*entity.Distribution, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	// Get distribution header with program and user info
	query := `
		SELECT d.id, d.distribution_date, d.program_id, p.id, p.name,
		       d.source_fund_type, d.total_amount, d.notes, d.created_by_user_id,
		       u.id, u.name, d.created_at, d.updated_at
		FROM distributions d
		LEFT JOIN programs p ON d.program_id = p.id
		INNER JOIN users u ON d.created_by_user_id = u.id
		WHERE d.id = $1
		LIMIT 1
	`

	d := &entity.Distribution{
		CreatedByUser: &entity.User{},
	}

	var programID, programName *string
	var distributionDate time.Time
	err := r.db.QueryRow(ctx, query, id).Scan(
		&d.ID, &distributionDate, &d.ProgramID, &programID, &programName,
		&d.SourceFundType, &d.TotalAmount, &d.Notes, &d.CreatedByUserID,
		&d.CreatedByUser.ID, &d.CreatedByUser.Name, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	// Convert time.Time to YYYY-MM-DD string
	d.DistributionDate = distributionDate.Format("2006-01-02")

	// Set program if exists
	if programID != nil && programName != nil {
		d.Program = &entity.Program{
			ID:   *programID,
			Name: *programName,
		}
	}

	// Get items with mustahiq and asnaf info
	itemsQuery := `
		SELECT di.id, di.distribution_id, di.mustahiq_id, m.name, a.name, m.address,
		       di.amount, di.notes, di.created_at, di.updated_at
		FROM distribution_items di
		INNER JOIN mustahiq m ON di.mustahiq_id = m.id
		INNER JOIN asnaf a ON m.asnafID = a.id
		WHERE di.distribution_id = $1
		ORDER BY di.created_at ASC
	`

	itemsRows, err := r.db.Query(ctx, itemsQuery, id)
	if err != nil {
		return nil, err
	}
	defer itemsRows.Close()

	var items []*entity.DistributionItem
	for itemsRows.Next() {
		item := &entity.DistributionItem{
			Mustahiq: &entity.Mustahiq{
				Asnaf: &entity.Asnaf{},
			},
		}
		err := itemsRows.Scan(
			&item.ID, &item.DistributionID, &item.MustahiqID,
			&item.Mustahiq.Name, &item.Mustahiq.Asnaf.Name, &item.Mustahiq.Address,
			&item.Amount, &item.Notes, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	d.Items = items
	return d, nil
}

func (r *DistributionRepository) Create(distribution *entity.Distribution) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	// Start transaction
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Insert distribution header
	distributionQuery := `
		INSERT INTO distributions (id, distribution_date, program_id, source_fund_type, total_amount, notes, created_by_user_id, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	err = tx.QueryRow(ctx, distributionQuery,
		distribution.DistributionDate, distribution.ProgramID, distribution.SourceFundType,
		distribution.TotalAmount, distribution.Notes, distribution.CreatedByUserID,
	).Scan(&distribution.ID, &distribution.CreatedAt, &distribution.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "foreign key") {
			return errors.New("program or user not found")
		}
		return err
	}

	// Insert items
	if len(distribution.Items) > 0 {
		itemQuery := `
			INSERT INTO distribution_items (id, distribution_id, mustahiq_id, amount, notes, created_at, updated_at)
			VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW(), NOW())
			RETURNING id, created_at, updated_at
		`

		for _, item := range distribution.Items {
			err = tx.QueryRow(ctx, itemQuery,
				distribution.ID, item.MustahiqID, item.Amount, item.Notes,
			).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
			if err != nil {
				if strings.Contains(err.Error(), "foreign key") {
					return errors.New("mustahiq not found")
				}
				return err
			}
			item.DistributionID = distribution.ID
		}
	}

	// Commit transaction
	return tx.Commit(ctx)
}

func (r *DistributionRepository) Update(distribution *entity.Distribution) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	// Start transaction
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Update distribution header
	distributionQuery := `
		UPDATE distributions
		SET distribution_date = $1, program_id = $2, source_fund_type = $3,
		    total_amount = $4, notes = $5, updated_at = NOW()
		WHERE id = $6
	`

	ct, err := tx.Exec(ctx, distributionQuery,
		distribution.DistributionDate, distribution.ProgramID, distribution.SourceFundType,
		distribution.TotalAmount, distribution.Notes, distribution.ID,
	)
	if err != nil {
		if strings.Contains(err.Error(), "foreign key") {
			return errors.New("program not found")
		}
		return err
	}

	if ct.RowsAffected() == 0 {
		return errors.New("distribution not found")
	}

	// Delete existing items
	_, err = tx.Exec(ctx, "DELETE FROM distribution_items WHERE distribution_id = $1", distribution.ID)
	if err != nil {
		return err
	}

	// Insert new items
	if len(distribution.Items) > 0 {
		itemQuery := `
			INSERT INTO distribution_items (id, distribution_id, mustahiq_id, amount, notes, created_at, updated_at)
			VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW(), NOW())
			RETURNING id, created_at, updated_at
		`

		for _, item := range distribution.Items {
			err = tx.QueryRow(ctx, itemQuery,
				distribution.ID, item.MustahiqID, item.Amount, item.Notes,
			).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
			if err != nil {
				if strings.Contains(err.Error(), "foreign key") {
					return errors.New("mustahiq not found")
				}
				return err
			}
			item.DistributionID = distribution.ID
		}
	}

	// Commit transaction
	return tx.Commit(ctx)
}

func (r *DistributionRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `DELETE FROM distributions WHERE id = $1`

	ct, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if ct.RowsAffected() == 0 {
		return errors.New("distribution not found")
	}

	return nil
}
