package usecase

import (
	"go-zakat-be/internal/domain/entity"
	"go-zakat-be/internal/domain/repository"

	"github.com/go-playground/validator/v10"
)

type AsnafUseCase struct {
	asnafRepo repository.AsnafRepository
	validator *validator.Validate
}

func NewAsnafUseCase(asnafRepo repository.AsnafRepository, validator *validator.Validate) *AsnafUseCase {
	return &AsnafUseCase{
		asnafRepo: asnafRepo,
		validator: validator,
	}
}

type CreateAsnafInput struct {
	Name        string `validate:"required"`
	Description string
}

type UpdateAsnafInput struct {
	ID          string `validate:"required"`
	Name        string `validate:"required"`
	Description string
}

func (uc *AsnafUseCase) Create(input CreateAsnafInput) (*entity.Asnaf, error) {
	if err := uc.validator.Struct(input); err != nil {
		return nil, err
	}

	asnaf := &entity.Asnaf{
		Name:        input.Name,
		Description: input.Description,
	}

	if err := uc.asnafRepo.Create(asnaf); err != nil {
		return nil, err
	}

	return asnaf, nil
}

func (uc *AsnafUseCase) FindAll(filter repository.AsnafFilter) ([]*entity.Asnaf, int64, error) {
	return uc.asnafRepo.FindAll(filter)
}

func (uc *AsnafUseCase) FindByID(id string) (*entity.Asnaf, error) {
	return uc.asnafRepo.FindByID(id)
}

func (uc *AsnafUseCase) Update(input UpdateAsnafInput) (*entity.Asnaf, error) {
	if err := uc.validator.Struct(input); err != nil {
		return nil, err
	}

	asnaf, err := uc.asnafRepo.FindByID(input.ID)
	if err != nil {
		return nil, err
	}

	asnaf.Name = input.Name
	asnaf.Description = input.Description

	if err := uc.asnafRepo.Update(asnaf); err != nil {
		return nil, err
	}

	return asnaf, nil
}

func (uc *AsnafUseCase) Delete(id string) error {
	return uc.asnafRepo.Delete(id)
}
