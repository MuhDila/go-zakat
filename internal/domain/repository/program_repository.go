package repository

import "go-zakat-be/internal/domain/entity"

type ProgramFilter struct {
	Query   string // Search by name
	Type    string // Filter by type
	Active  *bool  // Filter by active status (pointer to allow nil/true/false)
	Page    int
	PerPage int
}

type ProgramRepository interface {
	FindAll(filter ProgramFilter) ([]*entity.Program, int64, error)
	FindByID(id string) (*entity.Program, error)
	Create(program *entity.Program) error
	Update(program *entity.Program) error
	Delete(id string) error
}
