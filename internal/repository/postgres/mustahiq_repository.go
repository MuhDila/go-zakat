package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go-zakat/internal/domain/entity"
	"go-zakat/internal/domain/repository"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type MustahiqRepository struct {
	db  *pgxpool.Pool
	log *logrus.Logger
}

func NewMustahiqRepository(db *pgxpool.Pool, log *logrus.Logger) *MustahiqRepository {
	return &MustahiqRepository{db: db, log: log}
}

func (r *MustahiqRepository) FindAll(filter repository.MustahiqFilter) ([]*entity.Mustahiq, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	// Base query with JOIN to asnaf table
	query := `
		SELECT m.id, m.name, m.phoneNumber, m.address, m.asnafID, m.status, m.description, m.created_at, m.updated_at,
		       a.id as asnaf_id, a.name as asnaf_name
		FROM mustahiq m
		INNER JOIN asnaf a ON m.asnafID = a.id
	`
	countQuery := `SELECT COUNT(*) FROM mustahiq m INNER JOIN asnaf a ON m.asnafID = a.id`
	var args []interface{}
	argIdx := 1
	var conditions []string

	// Filter by query (name or address)
	if filter.Query != "" {
		search := fmt.Sprintf("%%%s%%", filter.Query)
		conditions = append(conditions, fmt.Sprintf("(m.name ILIKE $%d OR m.address ILIKE $%d)", argIdx, argIdx+1))
		args = append(args, search, search)
		argIdx += 2
	}

	// Filter by status
	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("m.status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}

	// Filter by asnaf ID
	if filter.AsnafID != "" {
		conditions = append(conditions, fmt.Sprintf("m.asnafID = $%d", argIdx))
		args = append(args, filter.AsnafID)
		argIdx++
	}

	// Add WHERE clause if there are conditions
	if len(conditions) > 0 {
		whereClause := " WHERE " + strings.Join(conditions, " AND ")
		query += whereClause
		countQuery += whereClause
	}

	// Get total count first
	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Pagination
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

	var mustahiqs []*entity.Mustahiq
	for rows.Next() {
		m := &entity.Mustahiq{
			Asnaf: &entity.Asnaf{}, // Initialize nested asnaf object
		}
		err := rows.Scan(
			&m.ID, &m.Name, &m.PhoneNumber, &m.Address, &m.AsnafID, &m.Status, &m.Description, &m.CreatedAt, &m.UpdatedAt,
			&m.Asnaf.ID, &m.Asnaf.Name,
		)
		if err != nil {
			return nil, 0, err
		}
		mustahiqs = append(mustahiqs, m)
	}

	return mustahiqs, total, nil
}

func (r *MustahiqRepository) FindByID(id string) (*entity.Mustahiq, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		SELECT m.id, m.name, m.phoneNumber, m.address, m.asnafID, m.status, m.description, m.created_at, m.updated_at,
		       a.id as asnaf_id, a.name as asnaf_name
		FROM mustahiq m
		INNER JOIN asnaf a ON m.asnafID = a.id
		WHERE m.id = $1
		LIMIT 1
	`

	m := &entity.Mustahiq{
		Asnaf: &entity.Asnaf{},
	}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&m.ID, &m.Name, &m.PhoneNumber, &m.Address, &m.AsnafID, &m.Status, &m.Description, &m.CreatedAt, &m.UpdatedAt,
		&m.Asnaf.ID, &m.Asnaf.Name,
	)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (r *MustahiqRepository) Create(mustahiq *entity.Mustahiq) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		INSERT INTO mustahiq (id, name, phoneNumber, address, asnafID, status, description, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query, mustahiq.Name, mustahiq.PhoneNumber, mustahiq.Address, mustahiq.AsnafID, mustahiq.Status, mustahiq.Description).
		Scan(&mustahiq.ID, &mustahiq.CreatedAt, &mustahiq.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return errors.New("nomor telepon sudah terdaftar")
		}
		if strings.Contains(err.Error(), "foreign key") {
			return errors.New("asnaf tidak ditemukan")
		}
		return err
	}

	return nil
}

func (r *MustahiqRepository) Update(mustahiq *entity.Mustahiq) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		UPDATE mustahiq
		SET name = $1, phoneNumber = $2, address = $3, asnafID = $4, status = $5, description = $6, updated_at = NOW()
		WHERE id = $7
	`

	ct, err := r.db.Exec(ctx, query, mustahiq.Name, mustahiq.PhoneNumber, mustahiq.Address, mustahiq.AsnafID, mustahiq.Status, mustahiq.Description, mustahiq.ID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return errors.New("nomor telepon sudah terdaftar")
		}
		if strings.Contains(err.Error(), "foreign key") {
			return errors.New("asnaf tidak ditemukan")
		}
		return err
	}

	if ct.RowsAffected() == 0 {
		return errors.New("mustahiq not found")
	}

	return nil
}

func (r *MustahiqRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `DELETE FROM mustahiq WHERE id = $1`

	ct, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if ct.RowsAffected() == 0 {
		return errors.New("mustahiq not found")
	}

	return nil
}
