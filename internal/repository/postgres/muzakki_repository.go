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

type MuzakkiRepository struct {
	db  *pgxpool.Pool
	log *logrus.Logger
}

func NewMuzakkiRepository(db *pgxpool.Pool, log *logrus.Logger) *MuzakkiRepository {
	return &MuzakkiRepository{db: db, log: log}
}

func (r *MuzakkiRepository) FindAll(filter repository.MuzakkiFilter) ([]*entity.Muzakki, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	// Base query
	query := `SELECT id, name, phoneNumber, address, notes, created_at, updated_at FROM muzakki`
	countQuery := `SELECT COUNT(*) FROM muzakki`
	var args []interface{}
	argIdx := 1

	// Filter by query (name or phone number)
	if filter.Query != "" {
		search := fmt.Sprintf("%%%s%%", filter.Query)
		condition := fmt.Sprintf(" WHERE (name ILIKE $%d OR phoneNumber ILIKE $%d)", argIdx, argIdx+1)
		query += condition
		countQuery += condition
		args = append(args, search, search)
		argIdx += 2
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

	var muzakkis []*entity.Muzakki
	for rows.Next() {
		m := &entity.Muzakki{}
		err := rows.Scan(&m.ID, &m.Name, &m.PhoneNumber, &m.Address, &m.Notes, &m.CreatedAt, &m.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		muzakkis = append(muzakkis, m)
	}

	return muzakkis, total, nil
}

func (r *MuzakkiRepository) FindByID(id string) (*entity.Muzakki, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		SELECT id, name, phoneNumber, address, notes, created_at, updated_at
		FROM muzakki
		WHERE id = $1
		LIMIT 1
	`

	m := &entity.Muzakki{}
	err := r.db.QueryRow(ctx, query, id).Scan(&m.ID, &m.Name, &m.PhoneNumber, &m.Address, &m.Notes, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (r *MuzakkiRepository) Create(muzakki *entity.Muzakki) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		INSERT INTO muzakki (id, name, phoneNumber, address, notes, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query, muzakki.Name, muzakki.PhoneNumber, muzakki.Address, muzakki.Notes).
		Scan(&muzakki.ID, &muzakki.CreatedAt, &muzakki.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return errors.New("nomor telepon sudah terdaftar")
		}
		return err
	}

	return nil
}

func (r *MuzakkiRepository) Update(muzakki *entity.Muzakki) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		UPDATE muzakki
		SET name = $1, phoneNumber = $2, address = $3, notes = $4, updated_at = NOW()
		WHERE id = $5
	`

	ct, err := r.db.Exec(ctx, query, muzakki.Name, muzakki.PhoneNumber, muzakki.Address, muzakki.Notes, muzakki.ID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return errors.New("nomor telepon sudah terdaftar")
		}
		return err
	}

	if ct.RowsAffected() == 0 {
		return errors.New("muzakki not found")
	}

	return nil
}

func (r *MuzakkiRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `DELETE FROM muzakki WHERE id = $1`

	ct, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if ct.RowsAffected() == 0 {
		return errors.New("muzakki not found")
	}

	return nil
}
