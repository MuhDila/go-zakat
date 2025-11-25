package repository

import "go-zakat/internal/domain/entity"

type AsnafFilter struct {
	Query   string
	Page    int
	PerPage int
}

type AsnafRepository interface {
	FindAll(filter AsnafFilter) ([]*entity.Asnaf, int64, error)
	FindByID(id string) (*entity.Asnaf, error)
	Create(asnaf *entity.Asnaf) error
	Update(asnaf *entity.Asnaf) error
	Delete(id string) error
}
