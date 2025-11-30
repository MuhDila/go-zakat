package usecase

import (
	"errors"
	"go-zakat-be/internal/domain/entity"
	"go-zakat-be/internal/domain/repository"

	"github.com/go-playground/validator/v10"
)

type UserUseCase struct {
	userRepo  repository.UserRepository
	validator *validator.Validate
}

func NewUserUseCase(userRepo repository.UserRepository, validator *validator.Validate) *UserUseCase {
	return &UserUseCase{
		userRepo:  userRepo,
		validator: validator,
	}
}

// FindAll returns list of users with pagination
func (uc *UserUseCase) FindAll(query, role string, page, perPage int) ([]*entity.User, int64, error) {
	// Validate pagination
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 10
	}

	// Validate role filter if provided
	if role != "" {
		validRoles := []string{entity.RoleAdmin, entity.RoleStaf, entity.RoleViewer}
		valid := false
		for _, r := range validRoles {
			if role == r {
				valid = true
				break
			}
		}
		if !valid {
			return nil, 0, errors.New("invalid role filter")
		}
	}

	filter := repository.UserFilter{
		Query:   query,
		Role:    role,
		Page:    page,
		PerPage: perPage,
	}

	return uc.userRepo.FindAll(filter)
}

// FindByID returns user by ID
func (uc *UserUseCase) FindByID(userID string) (*entity.User, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	return uc.userRepo.FindByID(userID)
}

// UpdateRole updates user role
func (uc *UserUseCase) UpdateRole(userID, role, currentUserID string) (*entity.User, error) {
	// Validate inputs
	if userID == "" {
		return nil, errors.New("user ID is required")
	}
	if role == "" {
		return nil, errors.New("role is required")
	}

	// Validate role value
	validRoles := []string{entity.RoleAdmin, entity.RoleStaf, entity.RoleViewer}
	valid := false
	for _, r := range validRoles {
		if role == r {
			valid = true
			break
		}
	}
	if !valid {
		return nil, errors.New("invalid role, must be: admin, staf, or viewer")
	}

	// Prevent admin from changing their own role
	if userID == currentUserID {
		return nil, errors.New("cannot change your own role")
	}

	// Check if user exists
	user, err := uc.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Update role
	err = uc.userRepo.UpdateRole(userID, role)
	if err != nil {
		return nil, err
	}

	// Return updated user
	user.Role = role
	return user, nil
}
