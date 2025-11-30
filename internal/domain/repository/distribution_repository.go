package repository

import "go-zakat-be/internal/domain/entity"

type DistributionFilter struct {
	DateFrom       string // YYYY-MM-DD
	DateTo         string // YYYY-MM-DD
	SourceFundType string // zakat_fitrah, zakat_maal, infaq, sadaqah
	ProgramID      string
	Query          string // search in program name or notes
	Page           int
	PerPage        int
}

type DistributionRepository interface {
	FindAll(filter DistributionFilter) ([]*entity.Distribution, int64, error)
	FindByID(id string) (*entity.Distribution, error)
	Create(distribution *entity.Distribution) error
	Update(distribution *entity.Distribution) error
	Delete(id string) error
}
