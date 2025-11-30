package usecase

import (
	"go-zakat-be/internal/domain/entity"
	"go-zakat-be/internal/domain/repository"

	"github.com/go-playground/validator/v10"
)

type MuzakkiUseCase struct {
	muzakkiRepo repository.MuzakkiRepository
	validator   *validator.Validate
}

func NewMuzakkiUseCase(muzakkiRepo repository.MuzakkiRepository, validator *validator.Validate) *MuzakkiUseCase {
	return &MuzakkiUseCase{
		muzakkiRepo: muzakkiRepo,
		validator:   validator,
	}
}

type CreateMuzakkiInput struct {
	Name        string `validate:"required"`
	PhoneNumber string `validate:"required"`
	Address     string `validate:"required"`
	Notes       string
}

type UpdateMuzakkiInput struct {
	ID          string `validate:"required"`
	Name        string `validate:"required"`
	PhoneNumber string `validate:"required"`
	Address     string `validate:"required"`
	Notes       string
}

func (uc *MuzakkiUseCase) Create(input CreateMuzakkiInput) (*entity.Muzakki, error) {
	if err := uc.validator.Struct(input); err != nil {
		return nil, err
	}

	muzakki := &entity.Muzakki{
		Name:        input.Name,
		PhoneNumber: input.PhoneNumber,
		Address:     input.Address,
		Notes:       input.Notes,
	}

	if err := uc.muzakkiRepo.Create(muzakki); err != nil {
		return nil, err
	}

	return muzakki, nil
}

func (uc *MuzakkiUseCase) FindAll(filter repository.MuzakkiFilter) ([]*entity.Muzakki, int64, error) {
	return uc.muzakkiRepo.FindAll(filter)
}

func (uc *MuzakkiUseCase) FindByID(id string) (*entity.Muzakki, error) {
	return uc.muzakkiRepo.FindByID(id)
}

func (uc *MuzakkiUseCase) Update(input UpdateMuzakkiInput) (*entity.Muzakki, error) {
	if err := uc.validator.Struct(input); err != nil {
		return nil, err
	}

	muzakki, err := uc.muzakkiRepo.FindByID(input.ID)
	if err != nil {
		return nil, err
	}

	muzakki.Name = input.Name
	muzakki.PhoneNumber = input.PhoneNumber
	muzakki.Address = input.Address
	muzakki.Notes = input.Notes

	if err := uc.muzakkiRepo.Update(muzakki); err != nil {
		return nil, err
	}

	return muzakki, nil
}

func (uc *MuzakkiUseCase) Delete(id string) error {
	return uc.muzakkiRepo.Delete(id)
}
