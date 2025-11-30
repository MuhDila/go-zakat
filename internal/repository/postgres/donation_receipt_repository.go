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

type DonationReceiptRepository struct {
	db  *pgxpool.Pool
	log *logrus.Logger
}

func NewDonationReceiptRepository(db *pgxpool.Pool, log *logrus.Logger) *DonationReceiptRepository {
	return &DonationReceiptRepository{db: db, log: log}
}

func (r *DonationReceiptRepository) FindAll(filter repository.DonationReceiptFilter) ([]*entity.DonationReceipt, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	// Base query with JOINs
	query := `
		SELECT DISTINCT dr.id, dr.receipt_number, dr.receipt_date, dr.muzakki_id, m.name as muzakki_name,
		       dr.payment_method, dr.total_amount, dr.notes, dr.created_by_user_id, dr.created_at, dr.updated_at
		FROM donation_receipts dr
		INNER JOIN muzakki m ON dr.muzakki_id = m.id
		LEFT JOIN donation_receipt_items dri ON dr.id = dri.receipt_id
	`

	countQuery := `
		SELECT COUNT(DISTINCT dr.id)
		FROM donation_receipts dr
		INNER JOIN muzakki m ON dr.muzakki_id = m.id
		LEFT JOIN donation_receipt_items dri ON dr.id = dri.receipt_id
	`

	var args []interface{}
	argIdx := 1
	var conditions []string

	// Filter by date range
	if filter.DateFrom != "" {
		conditions = append(conditions, fmt.Sprintf("dr.receipt_date >= $%d", argIdx))
		args = append(args, filter.DateFrom)
		argIdx++
	}
	if filter.DateTo != "" {
		conditions = append(conditions, fmt.Sprintf("dr.receipt_date <= $%d", argIdx))
		args = append(args, filter.DateTo)
		argIdx++
	}

	// Filter by fund_type (via items)
	if filter.FundType != "" {
		conditions = append(conditions, fmt.Sprintf("dri.fund_type = $%d", argIdx))
		args = append(args, filter.FundType)
		argIdx++
	}

	// Filter by zakat_type (via items)
	if filter.ZakatType != "" {
		conditions = append(conditions, fmt.Sprintf("dri.zakat_type = $%d", argIdx))
		args = append(args, filter.ZakatType)
		argIdx++
	}

	// Filter by payment_method
	if filter.PaymentMethod != "" {
		conditions = append(conditions, fmt.Sprintf("dr.payment_method = $%d", argIdx))
		args = append(args, filter.PaymentMethod)
		argIdx++
	}

	// Filter by muzakki_id
	if filter.MuzakkiID != "" {
		conditions = append(conditions, fmt.Sprintf("dr.muzakki_id = $%d", argIdx))
		args = append(args, filter.MuzakkiID)
		argIdx++
	}

	// Search in muzakki name or notes
	if filter.Query != "" {
		search := fmt.Sprintf("%%%s%%", filter.Query)
		conditions = append(conditions, fmt.Sprintf("(m.name ILIKE $%d OR dr.notes ILIKE $%d)", argIdx, argIdx+1))
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
	query += " ORDER BY dr.receipt_date DESC, dr.created_at DESC"
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

	var receipts []*entity.DonationReceipt
	for rows.Next() {
		dr := &entity.DonationReceipt{
			Muzakki: &entity.Muzakki{},
		}
		var receiptDate time.Time
		err := rows.Scan(
			&dr.ID, &dr.ReceiptNumber, &receiptDate, &dr.MuzakkiID, &dr.Muzakki.Name,
			&dr.PaymentMethod, &dr.TotalAmount, &dr.Notes, &dr.CreatedByUserID, &dr.CreatedAt, &dr.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		// Convert time.Time to YYYY-MM-DD string
		dr.ReceiptDate = receiptDate.Format("2006-01-02")
		receipts = append(receipts, dr)
	}

	return receipts, total, nil
}

func (r *DonationReceiptRepository) FindByID(id string) (*entity.DonationReceipt, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	// Get receipt header with muzakki and user info
	query := `
		SELECT dr.id, dr.receipt_number, dr.receipt_date, dr.muzakki_id, m.id, m.name,
		       dr.payment_method, dr.total_amount, dr.notes, dr.created_by_user_id,
		       u.id, u.name, dr.created_at, dr.updated_at
		FROM donation_receipts dr
		INNER JOIN muzakki m ON dr.muzakki_id = m.id
		INNER JOIN users u ON dr.created_by_user_id = u.id
		WHERE dr.id = $1
		LIMIT 1
	`

	dr := &entity.DonationReceipt{
		Muzakki:       &entity.Muzakki{},
		CreatedByUser: &entity.User{},
	}
	var receiptDate time.Time
	err := r.db.QueryRow(ctx, query, id).Scan(
		&dr.ID, &dr.ReceiptNumber, &receiptDate, &dr.MuzakkiID, &dr.Muzakki.ID, &dr.Muzakki.Name,
		&dr.PaymentMethod, &dr.TotalAmount, &dr.Notes, &dr.CreatedByUserID,
		&dr.CreatedByUser.ID, &dr.CreatedByUser.Name, &dr.CreatedAt, &dr.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	// Convert time.Time to YYYY-MM-DD string
	dr.ReceiptDate = receiptDate.Format("2006-01-02")

	// Get items
	itemsQuery := `
		SELECT id, receipt_id, fund_type, zakat_type, person_count, amount, rice_kg, notes, created_at, updated_at
		FROM donation_receipt_items
		WHERE receipt_id = $1
		ORDER BY created_at ASC
	`

	itemsRows, err := r.db.Query(ctx, itemsQuery, id)
	if err != nil {
		return nil, err
	}
	defer itemsRows.Close()

	var items []*entity.DonationReceiptItem
	for itemsRows.Next() {
		item := &entity.DonationReceiptItem{}
		err := itemsRows.Scan(
			&item.ID, &item.ReceiptID, &item.FundType, &item.ZakatType, &item.PersonCount,
			&item.Amount, &item.RiceKG, &item.Notes, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	dr.Items = items
	return dr, nil
}

func (r *DonationReceiptRepository) Create(receipt *entity.DonationReceipt) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	// Start transaction
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Insert receipt header
	receiptQuery := `
		INSERT INTO donation_receipts (id, muzakki_id, receipt_number, receipt_date, payment_method, total_amount, notes, created_by_user_id, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	err = tx.QueryRow(ctx, receiptQuery,
		receipt.MuzakkiID, receipt.ReceiptNumber, receipt.ReceiptDate, receipt.PaymentMethod,
		receipt.TotalAmount, receipt.Notes, receipt.CreatedByUserID,
	).Scan(&receipt.ID, &receipt.CreatedAt, &receipt.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return errors.New("receipt number already exists")
		}
		if strings.Contains(err.Error(), "foreign key") {
			return errors.New("muzakki or user not found")
		}
		return err
	}

	// Insert items
	if len(receipt.Items) > 0 {
		itemQuery := `
			INSERT INTO donation_receipt_items (id, receipt_id, fund_type, zakat_type, person_count, amount, rice_kg, notes, created_at, updated_at)
			VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
			RETURNING id, created_at, updated_at
		`

		for _, item := range receipt.Items {
			err = tx.QueryRow(ctx, itemQuery,
				receipt.ID, item.FundType, item.ZakatType, item.PersonCount,
				item.Amount, item.RiceKG, item.Notes,
			).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
			if err != nil {
				return err
			}
			item.ReceiptID = receipt.ID
		}
	}

	// Commit transaction
	return tx.Commit(ctx)
}

func (r *DonationReceiptRepository) Update(receipt *entity.DonationReceipt) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	// Start transaction
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Update receipt header
	receiptQuery := `
		UPDATE donation_receipts
		SET muzakki_id = $1, receipt_number = $2, receipt_date = $3, payment_method = $4,
		    total_amount = $5, notes = $6, updated_at = NOW()
		WHERE id = $7
	`

	ct, err := tx.Exec(ctx, receiptQuery,
		receipt.MuzakkiID, receipt.ReceiptNumber, receipt.ReceiptDate, receipt.PaymentMethod,
		receipt.TotalAmount, receipt.Notes, receipt.ID,
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return errors.New("receipt number already exists")
		}
		if strings.Contains(err.Error(), "foreign key") {
			return errors.New("muzakki not found")
		}
		return err
	}

	if ct.RowsAffected() == 0 {
		return errors.New("donation receipt not found")
	}

	// Delete existing items
	_, err = tx.Exec(ctx, "DELETE FROM donation_receipt_items WHERE receipt_id = $1", receipt.ID)
	if err != nil {
		return err
	}

	// Insert new items
	if len(receipt.Items) > 0 {
		itemQuery := `
			INSERT INTO donation_receipt_items (id, receipt_id, fund_type, zakat_type, person_count, amount, rice_kg, notes, created_at, updated_at)
			VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
			RETURNING id, created_at, updated_at
		`

		for _, item := range receipt.Items {
			err = tx.QueryRow(ctx, itemQuery,
				receipt.ID, item.FundType, item.ZakatType, item.PersonCount,
				item.Amount, item.RiceKG, item.Notes,
			).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
			if err != nil {
				return err
			}
			item.ReceiptID = receipt.ID
		}
	}

	// Commit transaction
	return tx.Commit(ctx)
}

func (r *DonationReceiptRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `DELETE FROM donation_receipts WHERE id = $1`

	ct, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if ct.RowsAffected() == 0 {
		return errors.New("donation receipt not found")
	}

	return nil
}
