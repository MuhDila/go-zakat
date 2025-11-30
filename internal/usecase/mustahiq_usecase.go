package usecase

import (
	"go-zakat-be/internal/domain/entity"
	"go-zakat-be/internal/domain/repository"

	"github.com/go-playground/validator/v10"
)

type MustahiqUseCase struct {
	mustahiqRepo repository.MustahiqRepository
	validator    *validator.Validate
}

func NewMustahiqUseCase(mustahiqRepo repository.MustahiqRepository, validator *validator.Validate) *MustahiqUseCase {
	return &MustahiqUseCase{
		mustahiqRepo: mustahiqRepo,
		validator:    validator,
	}
}

type CreateMustahiqInput struct {
	Name        string `validate:"required"`
	PhoneNumber string `validate:"required"`
	Address     string `validate:"required"`
	AsnafID     string `validate:"required"`
	Status      string `validate:"omitempty,oneof=active inactive pending"`
	Description string
}

type UpdateMustahiqInput struct {
	ID          string `validate:"required"`
	Name        string `validate:"required"`
	PhoneNumber string `validate:"required"`
	Address     string `validate:"required"`
	AsnafID     string `validate:"required"`
	Status      string `validate:"required,oneof=active inactive pending"`
	Description string
}

func (uc *MustahiqUseCase) Create(input CreateMustahiqInput) (*entity.Mustahiq, error) {
	if err := uc.validator.Struct(input); err != nil {
		return nil, err
	}

	// Set default status to pending if not provided
	status := input.Status
	if status == "" {
		status = entity.MustahiqStatusPending
	}

	mustahiq := &entity.Mustahiq{
		Name:        input.Name,
		PhoneNumber: input.PhoneNumber,
		Address:     input.Address,
		AsnafID:     input.AsnafID,
		Status:      status,
		Description: input.Description,
	}

	if err := uc.mustahiqRepo.Create(mustahiq); err != nil {
		return nil, err
	}

	return mustahiq, nil
}

func (uc *MustahiqUseCase) FindAll(filter repository.MustahiqFilter) ([]*entity.Mustahiq, int64, error) {
	return uc.mustahiqRepo.FindAll(filter)
}

func (uc *MustahiqUseCase) FindByID(id string) (*entity.Mustahiq, error) {
	return uc.mustahiqRepo.FindByID(id)
}

func (uc *MustahiqUseCase) Update(input UpdateMustahiqInput) (*entity.Mustahiq, error) {
	if err := uc.validator.Struct(input); err != nil {
		return nil, err
	}

	mustahiq, err := uc.mustahiqRepo.FindByID(input.ID)
	if err != nil {
		return nil, err
	}

	mustahiq.Name = input.Name
	mustahiq.PhoneNumber = input.PhoneNumber
	mustahiq.Address = input.Address
	mustahiq.AsnafID = input.AsnafID
	mustahiq.Status = input.Status
	mustahiq.Description = input.Description

	if err := uc.mustahiqRepo.Update(mustahiq); err != nil {
		return nil, err
	}

	return mustahiq, nil
}

func (uc *MustahiqUseCase) Delete(id string) error {
	return uc.mustahiqRepo.Delete(id)
}
