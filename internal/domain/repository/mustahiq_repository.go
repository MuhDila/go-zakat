package repository

import "go-zakat-be/internal/domain/entity"

type MustahiqFilter struct {
	Query   string // Search by name or address
	Status  string // Filter by status: active, inactive, pending
	AsnafID string // Filter by asnaf ID
	Page    int
	PerPage int
}

type MustahiqRepository interface {
	FindAll(filter MustahiqFilter) ([]*entity.Mustahiq, int64, error)
	FindByID(id string) (*entity.Mustahiq, error)
	Create(mustahiq *entity.Mustahiq) error
	Update(mustahiq *entity.Mustahiq) error
	Delete(id string) error
}
