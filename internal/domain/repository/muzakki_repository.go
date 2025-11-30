package repository

import "go-zakat-be/internal/domain/entity"

type MuzakkiFilter struct {
	Query   string
	Page    int
	PerPage int
}

type MuzakkiRepository interface {
	FindAll(filter MuzakkiFilter) ([]*entity.Muzakki, int64, error)
	FindByID(id string) (*entity.Muzakki, error)
	Create(muzakki *entity.Muzakki) error
	Update(muzakki *entity.Muzakki) error
	Delete(id string) error
}
