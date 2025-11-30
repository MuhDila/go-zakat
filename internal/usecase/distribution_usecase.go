package usecase

import (
	"errors"
	"go-zakat-be/internal/domain/entity"
	"go-zakat-be/internal/domain/repository"

	"github.com/go-playground/validator/v10"
)

type DistributionUseCase struct {
	distributionRepo repository.DistributionRepository
	mustahiqRepo     repository.MustahiqRepository
	validator        *validator.Validate
}

func NewDistributionUseCase(
	distributionRepo repository.DistributionRepository,
	mustahiqRepo repository.MustahiqRepository,
	validator *validator.Validate,
) *DistributionUseCase {
	return &DistributionUseCase{
		distributionRepo: distributionRepo,
		mustahiqRepo:     mustahiqRepo,
		validator:        validator,
	}
}

type CreateDistributionItemInput struct {
	MustahiqID string  `validate:"required"`
	Amount     float64 `validate:"required,gt=0"`
	Notes      string
}

type CreateDistributionInput struct {
	DistributionDate string  `validate:"required"` // YYYY-MM-DD
	ProgramID        *string // optional
	SourceFundType   string  `validate:"required,oneof=zakat_fitrah zakat_maal infaq sadaqah"`
	Notes            string
	CreatedByUserID  string                        `validate:"required"`
	Items            []CreateDistributionItemInput `validate:"required,min=1,dive"`
}

type UpdateDistributionInput struct {
	ID               string `validate:"required"`
	DistributionDate string `validate:"required"`
	ProgramID        *string
	SourceFundType   string `validate:"required,oneof=zakat_fitrah zakat_maal infaq sadaqah"`
	Notes            string
	Items            []CreateDistributionItemInput `validate:"required,min=1,dive"`
}

func (uc *DistributionUseCase) Create(input CreateDistributionInput) (*entity.Distribution, error) {
	// Validate input
	if err := uc.validator.Struct(input); err != nil {
		return nil, err
	}

	// Verify all mustahiq exist
	for _, item := range input.Items {
		_, err := uc.mustahiqRepo.FindByID(item.MustahiqID)
		if err != nil {
			return nil, errors.New("mustahiq not found: " + item.MustahiqID)
		}
	}

	// Calculate total amount
	var totalAmount float64
	items := make([]*entity.DistributionItem, len(input.Items))
	for i, itemInput := range input.Items {
		totalAmount += itemInput.Amount
		items[i] = &entity.DistributionItem{
			MustahiqID: itemInput.MustahiqID,
			Amount:     itemInput.Amount,
			Notes:      itemInput.Notes,
		}
	}

	distribution := &entity.Distribution{
		DistributionDate: input.DistributionDate,
		ProgramID:        input.ProgramID,
		SourceFundType:   input.SourceFundType,
		TotalAmount:      totalAmount,
		Notes:            input.Notes,
		CreatedByUserID:  input.CreatedByUserID,
		Items:            items,
	}

	if err := uc.distributionRepo.Create(distribution); err != nil {
		return nil, err
	}

	return distribution, nil
}

func (uc *DistributionUseCase) FindAll(filter repository.DistributionFilter) ([]*entity.Distribution, int64, error) {
	return uc.distributionRepo.FindAll(filter)
}

func (uc *DistributionUseCase) FindByID(id string) (*entity.Distribution, error) {
	return uc.distributionRepo.FindByID(id)
}

func (uc *DistributionUseCase) Update(input UpdateDistributionInput) (*entity.Distribution, error) {
	// Validate input
	if err := uc.validator.Struct(input); err != nil {
		return nil, err
	}

	// Verify distribution exists
	existing, err := uc.distributionRepo.FindByID(input.ID)
	if err != nil {
		return nil, errors.New("distribution not found")
	}

	// Verify all mustahiq exist
	for _, item := range input.Items {
		_, err := uc.mustahiqRepo.FindByID(item.MustahiqID)
		if err != nil {
			return nil, errors.New("mustahiq not found: " + item.MustahiqID)
		}
	}

	// Calculate total amount
	var totalAmount float64
	items := make([]*entity.DistributionItem, len(input.Items))
	for i, itemInput := range input.Items {
		totalAmount += itemInput.Amount
		items[i] = &entity.DistributionItem{
			MustahiqID: itemInput.MustahiqID,
			Amount:     itemInput.Amount,
			Notes:      itemInput.Notes,
		}
	}

	existing.DistributionDate = input.DistributionDate
	existing.ProgramID = input.ProgramID
	existing.SourceFundType = input.SourceFundType
	existing.TotalAmount = totalAmount
	existing.Notes = input.Notes
	existing.Items = items

	if err := uc.distributionRepo.Update(existing); err != nil {
		return nil, err
	}

	return existing, nil
}

func (uc *DistributionUseCase) Delete(id string) error {
	return uc.distributionRepo.Delete(id)
}
