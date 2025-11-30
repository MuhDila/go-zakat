package usecase

import (
	"go-zakat-be/internal/domain/entity"
	"go-zakat-be/internal/domain/repository"

	"github.com/go-playground/validator/v10"
)

type ProgramUseCase struct {
	programRepo repository.ProgramRepository
	validator   *validator.Validate
}

func NewProgramUseCase(programRepo repository.ProgramRepository, validator *validator.Validate) *ProgramUseCase {
	return &ProgramUseCase{
		programRepo: programRepo,
		validator:   validator,
	}
}

type CreateProgramInput struct {
	Name        string `validate:"required"`
	Type        string `validate:"required"`
	Description string
	Active      bool
}

type UpdateProgramInput struct {
	ID          string `validate:"required"`
	Name        string `validate:"required"`
	Type        string `validate:"required"`
	Description string
	Active      bool
}

func (uc *ProgramUseCase) Create(input CreateProgramInput) (*entity.Program, error) {
	if err := uc.validator.Struct(input); err != nil {
		return nil, err
	}

	program := &entity.Program{
		Name:        input.Name,
		Type:        input.Type,
		Description: input.Description,
		Active:      input.Active,
	}

	if err := uc.programRepo.Create(program); err != nil {
		return nil, err
	}

	return program, nil
}

func (uc *ProgramUseCase) FindAll(filter repository.ProgramFilter) ([]*entity.Program, int64, error) {
	return uc.programRepo.FindAll(filter)
}

func (uc *ProgramUseCase) FindByID(id string) (*entity.Program, error) {
	return uc.programRepo.FindByID(id)
}

func (uc *ProgramUseCase) Update(input UpdateProgramInput) (*entity.Program, error) {
	if err := uc.validator.Struct(input); err != nil {
		return nil, err
	}

	program, err := uc.programRepo.FindByID(input.ID)
	if err != nil {
		return nil, err
	}

	program.Name = input.Name
	program.Type = input.Type
	program.Description = input.Description
	program.Active = input.Active

	if err := uc.programRepo.Update(program); err != nil {
		return nil, err
	}

	return program, nil
}

func (uc *ProgramUseCase) Delete(id string) error {
	return uc.programRepo.Delete(id)
}
