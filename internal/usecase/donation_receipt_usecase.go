package usecase

import (
	"errors"
	"go-zakat-be/internal/domain/entity"
	"go-zakat-be/internal/domain/repository"

	"github.com/go-playground/validator/v10"
)

type DonationReceiptUseCase struct {
	receiptRepo repository.DonationReceiptRepository
	muzakkiRepo repository.MuzakkiRepository
	validator   *validator.Validate
}

func NewDonationReceiptUseCase(
	receiptRepo repository.DonationReceiptRepository,
	muzakkiRepo repository.MuzakkiRepository,
	validator *validator.Validate,
) *DonationReceiptUseCase {
	return &DonationReceiptUseCase{
		receiptRepo: receiptRepo,
		muzakkiRepo: muzakkiRepo,
		validator:   validator,
	}
}

type CreateDonationReceiptItemInput struct {
	FundType    string   `validate:"required,oneof=zakat infaq sadaqah"`
	ZakatType   *string  `validate:"omitempty,oneof=fitrah maal"`
	PersonCount *int     `validate:"omitempty,min=1"`
	Amount      float64  `validate:"required,gt=0"`
	RiceKG      *float64 `validate:"omitempty,gt=0"`
	Notes       string
}

type CreateDonationReceiptInput struct {
	MuzakkiID       string `validate:"required"`
	ReceiptNumber   string `validate:"required"`
	ReceiptDate     string `validate:"required"` // YYYY-MM-DD
	PaymentMethod   string `validate:"required"`
	Notes           string
	CreatedByUserID string                           `validate:"required"`
	Items           []CreateDonationReceiptItemInput `validate:"required,min=1,dive"`
}

type UpdateDonationReceiptInput struct {
	ID            string `validate:"required"`
	MuzakkiID     string `validate:"required"`
	ReceiptNumber string `validate:"required"`
	ReceiptDate   string `validate:"required"`
	PaymentMethod string `validate:"required"`
	Notes         string
	Items         []CreateDonationReceiptItemInput `validate:"required,min=1,dive"`
}

func (uc *DonationReceiptUseCase) Create(input CreateDonationReceiptInput) (*entity.DonationReceipt, error) {
	// Validate input
	if err := uc.validator.Struct(input); err != nil {
		return nil, err
	}

	// Additional validation: if fund_type = zakat, zakat_type is required
	for i, item := range input.Items {
		if item.FundType == "zakat" && (item.ZakatType == nil || *item.ZakatType == "") {
			return nil, errors.New("zakat_type is required when fund_type is zakat (item " + string(rune(i+1)) + ")")
		}
	}

	// Verify muzakki exists
	_, err := uc.muzakkiRepo.FindByID(input.MuzakkiID)
	if err != nil {
		return nil, errors.New("muzakki not found")
	}

	// Calculate total amount
	var totalAmount float64
	items := make([]*entity.DonationReceiptItem, len(input.Items))
	for i, itemInput := range input.Items {
		totalAmount += itemInput.Amount
		items[i] = &entity.DonationReceiptItem{
			FundType:    itemInput.FundType,
			ZakatType:   itemInput.ZakatType,
			PersonCount: itemInput.PersonCount,
			Amount:      itemInput.Amount,
			RiceKG:      itemInput.RiceKG,
			Notes:       itemInput.Notes,
		}
	}

	receipt := &entity.DonationReceipt{
		MuzakkiID:       input.MuzakkiID,
		ReceiptNumber:   input.ReceiptNumber,
		ReceiptDate:     input.ReceiptDate,
		PaymentMethod:   input.PaymentMethod,
		TotalAmount:     totalAmount,
		Notes:           input.Notes,
		CreatedByUserID: input.CreatedByUserID,
		Items:           items,
	}

	if err := uc.receiptRepo.Create(receipt); err != nil {
		return nil, err
	}

	return receipt, nil
}

func (uc *DonationReceiptUseCase) FindAll(filter repository.DonationReceiptFilter) ([]*entity.DonationReceipt, int64, error) {
	return uc.receiptRepo.FindAll(filter)
}

func (uc *DonationReceiptUseCase) FindByID(id string) (*entity.DonationReceipt, error) {
	return uc.receiptRepo.FindByID(id)
}

func (uc *DonationReceiptUseCase) Update(input UpdateDonationReceiptInput) (*entity.DonationReceipt, error) {
	// Validate input
	if err := uc.validator.Struct(input); err != nil {
		return nil, err
	}

	// Additional validation
	for i, item := range input.Items {
		if item.FundType == "zakat" && (item.ZakatType == nil || *item.ZakatType == "") {
			return nil, errors.New("zakat_type is required when fund_type is zakat (item " + string(rune(i+1)) + ")")
		}
	}

	// Verify receipt exists
	existing, err := uc.receiptRepo.FindByID(input.ID)
	if err != nil {
		return nil, errors.New("donation receipt not found")
	}

	// Verify muzakki exists
	_, err = uc.muzakkiRepo.FindByID(input.MuzakkiID)
	if err != nil {
		return nil, errors.New("muzakki not found")
	}

	// Calculate total amount
	var totalAmount float64
	items := make([]*entity.DonationReceiptItem, len(input.Items))
	for i, itemInput := range input.Items {
		totalAmount += itemInput.Amount
		items[i] = &entity.DonationReceiptItem{
			FundType:    itemInput.FundType,
			ZakatType:   itemInput.ZakatType,
			PersonCount: itemInput.PersonCount,
			Amount:      itemInput.Amount,
			RiceKG:      itemInput.RiceKG,
			Notes:       itemInput.Notes,
		}
	}

	existing.MuzakkiID = input.MuzakkiID
	existing.ReceiptNumber = input.ReceiptNumber
	existing.ReceiptDate = input.ReceiptDate
	existing.PaymentMethod = input.PaymentMethod
	existing.TotalAmount = totalAmount
	existing.Notes = input.Notes
	existing.Items = items

	if err := uc.receiptRepo.Update(existing); err != nil {
		return nil, err
	}

	return existing, nil
}

func (uc *DonationReceiptUseCase) Delete(id string) error {
	return uc.receiptRepo.Delete(id)
}
