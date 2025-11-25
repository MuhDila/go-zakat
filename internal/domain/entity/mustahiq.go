package entity

import "time"

type Mustahiq struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	PhoneNumber string    `json:"phoneNumber"`
	Address     string    `json:"address"`
	AsnafID     string    `json:"asnafID"`
	Asnaf       *Asnaf    `json:"asnaf,omitempty"` // Nested asnaf object
	Status      string    `json:"status"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
