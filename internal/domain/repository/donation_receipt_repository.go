package repository

import "go-zakat-be/internal/domain/entity"

type DonationReceiptFilter struct {
	DateFrom      string // YYYY-MM-DD
	DateTo        string // YYYY-MM-DD
	FundType      string // zakat, infaq, sadaqah (filter by item's fund_type)
	ZakatType     string // fitrah, maal
	PaymentMethod string
	MuzakkiID     string
	Query         string // search in muzakki.full_name or notes
	Page          int
	PerPage       int
}

type DonationReceiptRepository interface {
	FindAll(filter DonationReceiptFilter) ([]*entity.DonationReceipt, int64, error)
	FindByID(id string) (*entity.DonationReceipt, error)
	Create(receipt *entity.DonationReceipt) error
	Update(receipt *entity.DonationReceipt) error
	Delete(id string) error
}
