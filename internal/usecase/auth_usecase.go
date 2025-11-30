package usecase

import (
	"errors"
	"fmt"

	"go-zakat-be/internal/domain/entity"
	"go-zakat-be/internal/domain/repository"
	"go-zakat-be/internal/domain/service"

	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

// AuthUseCase menyimpan dependency yang dibutuhkan oleh fitur auth
type AuthUseCase struct {
	userRepo  repository.UserRepository
	tokenSvc  service.TokenService
	googleSvc service.GoogleOAuthService
	validator *validator.Validate
}

// NewAuthUseCase membuat instance AuthUseCase
func NewAuthUseCase(
	userRepo repository.UserRepository,
	tokenSvc service.TokenService,
	googleSvc service.GoogleOAuthService,
	val *validator.Validate,
) *AuthUseCase {
	return &AuthUseCase{
		userRepo:  userRepo,
		tokenSvc:  tokenSvc,
		googleSvc: googleSvc,
		validator: val,
	}
}

// RegisterInput adalah bentuk input untuk register di layer usecase
type RegisterInput struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=6"`
	Name     string `validate:"required"`
}

type LoginInput struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
}

// AuthTokens output token
type AuthTokens struct {
	AccessToken  string
	RefreshToken string
}

// Register melakukan proses register user baru
func (uc *AuthUseCase) Register(input RegisterInput) (*AuthTokens, *entity.User, error) {
	// 1. Validasi input pakai validator
	if err := uc.validator.Struct(input); err != nil {
		return nil, nil, err
	}

	// 2. Cek apakah email sudah digunakan
	_, err := uc.userRepo.FindByEmail(input.Email)
	if err == nil {
		// kalau tidak error artinya user ada
		return nil, nil, errors.New("email sudah terdaftar")
	}

	// 3. Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
	if err != nil {
		return nil, nil, err
	}

	// 4. Buat entity User
	user := &entity.User{
		Email:    input.Email,
		Password: string(hashed),
		Name:     input.Name,
		Role:     entity.RoleViewer, // Default role
	}

	// 5. Simpan ke DB via UserRepository
	if err := uc.userRepo.Create(user); err != nil {
		return nil, nil, err
	}

	// 6. Generate access token & refresh token
	access, err := uc.tokenSvc.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, nil, err
	}
	refresh, err := uc.tokenSvc.GenerateRefreshToken(user.ID, user.Role)
	if err != nil {
		return nil, nil, err
	}

	return &AuthTokens{
		AccessToken:  access,
		RefreshToken: refresh,
	}, user, nil
}

// Login melakukan proses login
func (uc *AuthUseCase) Login(input LoginInput) (*AuthTokens, *entity.User, error) {
	if err := uc.validator.Struct(input); err != nil {
		return nil, nil, err
	}

	user, err := uc.userRepo.FindByEmail(input.Email)
	if err != nil {
		return nil, nil, errors.New("email atau password salah")
	}

	// compare password plaintext dengan hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return nil, nil, errors.New("email atau password salah")
	}

	access, err := uc.tokenSvc.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, nil, err
	}
	refresh, err := uc.tokenSvc.GenerateRefreshToken(user.ID, user.Role)
	if err != nil {
		return nil, nil, err
	}

	return &AuthTokens{
		AccessToken:  access,
		RefreshToken: refresh,
	}, user, nil
}

// GoogleLogin hanya mengembalikan URL untuk redirect (web)
func (uc *AuthUseCase) GoogleLogin(state string) (string, error) {
	// state sebaiknya disimpan di session/redis untuk validasi saat callback
	return uc.googleSvc.GetAuthURL(state), nil
}

// GoogleCallback memproses code dari Google dan generate token
func (uc *AuthUseCase) GoogleCallback(state, expectedState, code string) (*AuthTokens, *entity.User, error) {
	// 1. Validasi state (CSRF protection)
	if state != expectedState {
		return nil, nil, errors.New("state tidak valid")
	}

	// 2. Tukar code dengan access token
	accessToken, err := uc.googleSvc.ExchangeCode(code)
	if err != nil {
		return nil, nil, err
	}

	// 3. Ambil user info dari Google
	email, name, googleID, err := uc.googleSvc.GetUserInfo(accessToken)
	if err != nil {
		return nil, nil, err
	}

	// 4. Cek apakah user dengan google_id sudah ada
	user, err := uc.userRepo.FindByGoogleID(googleID)
	if err != nil {
		// asumsi err berarti belum ada → create user baru
		user = &entity.User{
			Email: email,
			Name:  name,
			Role:  entity.RoleViewer, // Default role
		}
		user.GoogleID = &googleID

		if err := uc.userRepo.Create(user); err != nil {
			return nil, nil, err
		}
	}

	// 5. Generate token
	access, err := uc.tokenSvc.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, nil, err
	}
	refresh, err := uc.tokenSvc.GenerateRefreshToken(user.ID, user.Role)
	if err != nil {
		return nil, nil, err
	}

	return &AuthTokens{
		AccessToken:  access,
		RefreshToken: refresh,
	}, user, nil
}

// RefreshToken : validasi refresh token → buat access token baru
func (uc *AuthUseCase) RefreshToken(refreshToken string) (*AuthTokens, error) {
	userID, _, err := uc.tokenSvc.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, errors.New("refresh token tidak valid")
	}

	// Ambil data user terbaru dari DB untuk memastikan role update
	user, err := uc.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user tidak ditemukan")
	}

	// Generate access token dengan role terbaru dari DB
	access, err := uc.tokenSvc.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	return &AuthTokens{
		AccessToken:  access,
		RefreshToken: refreshToken, // refresh token tetap sama (simple version)
	}, nil
}

func (uc *AuthUseCase) GetUserByID(userID string) (*entity.User, error) {
	user, err := uc.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user tidak ditemukan")
	}

	return user, nil
}

func (uc *AuthUseCase) GoogleMobileLogin(idToken string) (*AuthTokens, *entity.User, error) {
	email, name, googleID, err := uc.googleSvc.VerifyMobileIDToken(idToken)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid_google_token")
	}

	user, err := uc.userRepo.FindByGoogleID(googleID)
	if err != nil {
		// user belum ada → buat baru
		newUser := &entity.User{
			Email:    email,
			Name:     name,
			GoogleID: &googleID,
			Role:     entity.RoleViewer, // Default role
		}
		if err := uc.userRepo.Create(newUser); err != nil {
			return nil, nil, err
		}
		user = newUser
	}

	access, err := uc.tokenSvc.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, nil, err
	}

	refresh, err := uc.tokenSvc.GenerateRefreshToken(user.ID, user.Role)
	if err != nil {
		return nil, nil, err
	}

	return &AuthTokens{
		AccessToken:  access,
		RefreshToken: refresh,
	}, user, nil
}
