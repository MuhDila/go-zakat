package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go-zakat-be/internal/domain/entity"
	"go-zakat-be/internal/domain/repository"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type AsnafRepository struct {
	db  *pgxpool.Pool
	log *logrus.Logger
}

func NewAsnafRepository(db *pgxpool.Pool, log *logrus.Logger) *AsnafRepository {
	return &AsnafRepository{db: db, log: log}
}

func (r *AsnafRepository) FindAll(filter repository.AsnafFilter) ([]*entity.Asnaf, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	// Base query
	query := `SELECT id, name, description, created_at, updated_at FROM asnaf`
	countQuery := `SELECT COUNT(*) FROM asnaf`
	var args []interface{}
	argIdx := 1

	// Filter by query (name)
	if filter.Query != "" {
		search := fmt.Sprintf("%%%s%%", filter.Query)
		condition := fmt.Sprintf(" WHERE name ILIKE $%d", argIdx)
		query += condition
		countQuery += condition
		args = append(args, search)
		argIdx++
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

	var asnafs []*entity.Asnaf
	for rows.Next() {
		a := &entity.Asnaf{}
		err := rows.Scan(&a.ID, &a.Name, &a.Description, &a.CreatedAt, &a.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		asnafs = append(asnafs, a)
	}

	return asnafs, total, nil
}

func (r *AsnafRepository) FindByID(id string) (*entity.Asnaf, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		SELECT id, name, description, created_at, updated_at
		FROM asnaf
		WHERE id = $1
		LIMIT 1
	`

	a := &entity.Asnaf{}
	err := r.db.QueryRow(ctx, query, id).Scan(&a.ID, &a.Name, &a.Description, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func (r *AsnafRepository) Create(asnaf *entity.Asnaf) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		INSERT INTO asnaf (id, name, description, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query, asnaf.Name, asnaf.Description).
		Scan(&asnaf.ID, &asnaf.CreatedAt, &asnaf.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return errors.New("nama asnaf sudah terdaftar")
		}
		return err
	}

	return nil
}

func (r *AsnafRepository) Update(asnaf *entity.Asnaf) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		UPDATE asnaf
		SET name = $1, description = $2, updated_at = NOW()
		WHERE id = $3
	`

	ct, err := r.db.Exec(ctx, query, asnaf.Name, asnaf.Description, asnaf.ID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return errors.New("nama asnaf sudah terdaftar")
		}
		return err
	}

	if ct.RowsAffected() == 0 {
		return errors.New("asnaf not found")
	}

	return nil
}

func (r *AsnafRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `DELETE FROM asnaf WHERE id = $1`

	ct, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if ct.RowsAffected() == 0 {
		return errors.New("asnaf not found")
	}

	return nil
}
