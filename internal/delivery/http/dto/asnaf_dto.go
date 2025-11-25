package dto

import "time"

type CreateAsnafRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type UpdateAsnafRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type AsnafResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
