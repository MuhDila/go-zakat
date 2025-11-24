package entity

import "time"

const (
	RoleAdmin  = "admin"
	RoleStaff  = "staff"
	RoleViewer = "viewer"
)

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	GoogleID  *string   `json:"google_id"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
