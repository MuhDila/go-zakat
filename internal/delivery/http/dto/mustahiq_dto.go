package dto

import "time"

type CreateMustahiqRequest struct {
	Name        string `json:"name" binding:"required"`
	PhoneNumber string `json:"phoneNumber" binding:"required"`
	Address     string `json:"address" binding:"required"`
	AsnafID     string `json:"asnafID" binding:"required"`
	Status      string `json:"status" binding:"required,oneof=active inactive pending"`
	Description string `json:"description"`
}

type UpdateMustahiqRequest struct {
	Name        string `json:"name" binding:"required"`
	PhoneNumber string `json:"phoneNumber" binding:"required"`
	Address     string `json:"address" binding:"required"`
	AsnafID     string `json:"asnafID" binding:"required"`
	Status      string `json:"status" binding:"required,oneof=active inactive pending"`
	Description string `json:"description"`
}

type AsnafInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type MustahiqResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	PhoneNumber string    `json:"phoneNumber"`
	Address     string    `json:"address"`
	Asnaf       AsnafInfo `json:"asnaf"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
