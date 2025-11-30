package repository

import "go-zakat-be/internal/domain/entity"

type UserFilter struct {
	Query   string // Search in name or email
	Role    string // Filter by role
	Page    int
	PerPage int
}

type UserRepository interface {
	Create(user *entity.User) error
	FindByEmail(email string) (*entity.User, error)
	FindByID(id string) (*entity.User, error)
	FindByGoogleID(googleID string) (*entity.User, error)
	Update(user *entity.User) error
	FindAll(filter UserFilter) ([]*entity.User, int64, error)
	UpdateRole(userID, role string) error
}
