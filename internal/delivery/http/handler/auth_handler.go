package handler

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"go-zakat/internal/delivery/http/dto"
	"go-zakat/internal/usecase"
	"go-zakat/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authUC *usecase.AuthUseCase
}

func NewAuthHandler(authUC *usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{authUC: authUC}
}

// Register godoc
// @Summary Register user baru
// @Description Mendaftarkan user baru menggunakan email & password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Register Request Body"
// @Success 201 {object} dto.AuthResponseWrapper
// @Failure 400 {object} dto.ErrorResponseWrapper
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, gin.H{"error": err.Error()})
		return
	}

	tokens, user, err := h.authUC.Register(usecase.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	})
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Success(c, http.StatusCreated, "Register successful", dto.AuthResponse{
		User: dto.UserResponse{
			ID:    user.ID,
			Email: user.Email,
			Name:  user.Name,
			Role:  user.Role,
		},
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}

// Login godoc
// @Summary Login user
// @Description Login dengan email dan password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login Body"
// @Success 200 {object} dto.AuthResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, gin.H{"error": err.Error()})
		return
	}

	tokens, user, err := h.authUC.Login(usecase.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		response.Unauthorized(c, err.Error(), nil)
		return
	}

	response.Success(c, http.StatusOK, "Login successful", dto.AuthResponse{
		User: dto.UserResponse{
			ID:    user.ID,
			Email: user.Email,
			Name:  user.Name,
			Role:  user.Role,
		},
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}

// Me godoc
// @Summary Get data user yang sedang login
// @Description Mengambil informasi user berdasarkan access token yang dikirim di header
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.UserResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Router /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		response.Unauthorized(c, "user_id tidak ditemukan di context", nil)
		return
	}

	// Ambil user dari UseCase
	user, err := h.authUC.GetUserByID(userID.(string))
	if err != nil {
		response.Unauthorized(c, "User tidak ditemukan", nil)
		return
	}

	// Mapping ke response DTO
	response.Success(c, http.StatusOK, "Get user successful", dto.UserResponse{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
		Role:  user.Role,
	})
}

// Refresh godoc
// @Summary Refresh access token
// @Description Mengambil access token baru menggunakan refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh Token Body"
// @Success 200 {object} dto.AuthTokensResponseWrapper
// @Failure 400 {object} dto.ErrorResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshTokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, gin.H{"error": err.Error()})
		return
	}

	tokens, err := h.authUC.RefreshToken(req.RefreshToken)
	if err != nil {
		response.Unauthorized(c, err.Error(), nil)
		return
	}

	response.Success(c, http.StatusOK, "Refresh token successful", dto.AuthTokensResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}

// GoogleLogin godoc
// @Summary Get Google OAuth URL
// @Description Mengembalikan URL untuk redirect user ke Google OAuth
// @Tags Auth
// @Produce json
// @Success 200 {object} dto.AuthURLResponseWrapper
// @Failure 500 {object} dto.ErrorResponseWrapper
// @Router /auth/google/login [get]
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	// 1. Generate state random
	state, err := generateState()
	if err != nil {
		response.InternalServerError(c, "gagal generate state", nil)
		return
	}

	// 2. Simpan state di cookie (sederhana, untuk demo)
	// Di production sebaiknya pakai Redis/session store
	c.SetCookie("oauth_state", state, 300, "/", "", false, true) // 5 menit

	// 3. Minta URL ke UseCase
	authURL, err := h.authUC.GoogleLogin(state)
	if err != nil {
		response.InternalServerError(c, err.Error(), nil)
		return
	}

	// Bisa juga langsung redirect:
	// c.Redirect(http.StatusFound, authURL); return

	// Untuk demo/frontend, enak dikirim JSON
	response.Success(c, http.StatusOK, "Get Google login URL successful", dto.AuthURLResponse{
		AuthURL: authURL,
	})
}

// GoogleCallback godoc
// @Summary Google OAuth callback
// @Description Callback endpoint yang dipanggil oleh Google setelah user login
// @Tags Auth
// @Produce json
// @Param code query string true "Kode authorization dari Google"
// @Param state query string true "State untuk CSRF protection"
// @Success 200 {object} dto.AuthResponseWrapper
// @Failure 400 {object} dto.ErrorResponseWrapper
// @Failure 401 {object} dto.ErrorResponseWrapper
// @Failure 500 {object} dto.ErrorResponseWrapper
// @Router /auth/google/callback [get]
func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	// 1. Ambil code & state dari query
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		response.BadRequest(c, "code atau state kosong", nil)
		return
	}

	// 2. Ambil expectedState dari cookie
	expectedState, err := c.Cookie("oauth_state")
	if err != nil {
		response.Unauthorized(c, "state cookie tidak ditemukan", nil)
		return
	}

	// 3. Panggil UseCase
	tokens, user, err := h.authUC.GoogleCallback(state, expectedState, code)
	if err != nil {
		response.Unauthorized(c, err.Error(), nil)
		return
	}

	// 4. Beres, balikin token
	response.Success(c, http.StatusOK, "Google login successful", dto.AuthResponse{
		User: dto.UserResponse{
			ID:    user.ID,
			Email: user.Email,
			Name:  user.Name,
			Role:  user.Role,
		},
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}

// GoogleMobileLogin godoc
// @Summary Login dengan Google untuk aplikasi mobile (native)
// @Description Menerima id_token dari Google (hasil dari SDK Google di mobile), memverifikasi ke Google, membuat/mencari user di DB, lalu mengembalikan JWT access & refresh token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.GoogleMobileLoginRequest true "Body berisi id_token dari Google"
// @Success 200 {object} dto.AuthResponseWrapper "Berhasil login dengan Google (mobile)"
// @Failure 400 {object} dto.ErrorResponseWrapper "Body request tidak valid"
// @Failure 401 {object} dto.ErrorResponseWrapper "id_token Google tidak valid atau tidak bisa diverifikasi"
// @Router /auth/google/mobile/login [post]
func (h *AuthHandler) GoogleMobileLogin(c *gin.Context) {
	var req dto.GoogleMobileLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, gin.H{"error": err.Error()})
		return
	}

	tokens, user, err := h.authUC.GoogleMobileLogin(req.IDToken)
	if err != nil {
		response.Unauthorized(c, err.Error(), nil)
		return
	}

	response.Success(c, http.StatusOK, "Google mobile login successful", dto.AuthResponse{
		User: dto.UserResponse{
			ID:    user.ID,
			Email: user.Email,
			Name:  user.Name,
			Role:  user.Role,
		},
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}

// Helper
func generateState() (string, error) {
	b := make([]byte, 16) // 128-bit random
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
