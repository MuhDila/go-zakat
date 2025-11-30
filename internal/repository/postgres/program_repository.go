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

type ProgramRepository struct {
	db  *pgxpool.Pool
	log *logrus.Logger
}

func NewProgramRepository(db *pgxpool.Pool, log *logrus.Logger) *ProgramRepository {
	return &ProgramRepository{db: db, log: log}
}

func (r *ProgramRepository) FindAll(filter repository.ProgramFilter) ([]*entity.Program, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	// Base query
	query := `SELECT id, name, type, description, active, created_at, updated_at FROM programs`
	countQuery := `SELECT COUNT(*) FROM programs`
	var args []interface{}
	argIdx := 1
	var conditions []string

	// Filter by query (name)
	if filter.Query != "" {
		search := fmt.Sprintf("%%%s%%", filter.Query)
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argIdx))
		args = append(args, search)
		argIdx++
	}

	// Filter by type
	if filter.Type != "" {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argIdx))
		args = append(args, filter.Type)
		argIdx++
	}

	// Filter by active status
	if filter.Active != nil {
		conditions = append(conditions, fmt.Sprintf("active = $%d", argIdx))
		args = append(args, *filter.Active)
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

	var programs []*entity.Program
	for rows.Next() {
		p := &entity.Program{}
		err := rows.Scan(&p.ID, &p.Name, &p.Type, &p.Description, &p.Active, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		programs = append(programs, p)
	}

	return programs, total, nil
}

func (r *ProgramRepository) FindByID(id string) (*entity.Program, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		SELECT id, name, type, description, active, created_at, updated_at
		FROM programs
		WHERE id = $1
		LIMIT 1
	`

	p := &entity.Program{}
	err := r.db.QueryRow(ctx, query, id).Scan(&p.ID, &p.Name, &p.Type, &p.Description, &p.Active, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (r *ProgramRepository) Create(program *entity.Program) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		INSERT INTO programs (id, name, type, description, active, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query, program.Name, program.Type, program.Description, program.Active).
		Scan(&program.ID, &program.CreatedAt, &program.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return errors.New("nama program sudah terdaftar")
		}
		return err
	}

	return nil
}

func (r *ProgramRepository) Update(program *entity.Program) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		UPDATE programs
		SET name = $1, type = $2, description = $3, active = $4, updated_at = NOW()
		WHERE id = $5
	`

	ct, err := r.db.Exec(ctx, query, program.Name, program.Type, program.Description, program.Active, program.ID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return errors.New("nama program sudah terdaftar")
		}
		return err
	}

	if ct.RowsAffected() == 0 {
		return errors.New("program not found")
	}

	return nil
}

func (r *ProgramRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `DELETE FROM programs WHERE id = $1`

	ct, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if ct.RowsAffected() == 0 {
		return errors.New("program not found")
	}

	return nil
}
