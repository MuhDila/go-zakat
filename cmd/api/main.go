// @title Auth API
// @version 1.0
// @description API untuk autentikasi (register, login, refresh token, Google OAuth)
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

package main

import (
	"context"
	"errors"
	"go-zakat-be/pkg/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"go-zakat-be/docs"

	"go-zakat-be/internal/delivery/http/handler"
	"go-zakat-be/internal/delivery/http/middleware"
	domainValidator "go-zakat-be/internal/delivery/http/validator"
	"go-zakat-be/internal/infrastructure/jwt"
	"go-zakat-be/internal/infrastructure/oauth"
	"go-zakat-be/internal/repository/postgres"
	"go-zakat-be/internal/usecase"

	"go-zakat-be/pkg/config"
	"go-zakat-be/pkg/database"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()
	logr := logger.New() // <-- logger init

	// Swagger info
	docs.SwaggerInfo.Title = "Auth API"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:" + cfg.AppPort

	// Run Migrations
	if err := database.RunMigrations(cfg.DatabaseURL); err != nil {
		logr.Fatalf("gagal run migrations: %v", err)
	}

	// Database
	dbPool, err := database.NewPostgresPool(cfg.DatabaseURL)
	if err != nil {
		logr.Fatalf("gagal init DB: %v", err)
	}
	defer dbPool.Close()

	val := domainValidator.NewValidator()

	// JWT
	tokenCfg := jwt.TokenConfig{
		AccessSecret:    cfg.JWTAccessSecret,
		RefreshSecret:   cfg.JWTRefreshSecret,
		AccessTokenTTL:  cfg.JWTAccessTTL,
		RefreshTokenTTL: cfg.JWTRefreshTTL,
	}
	tokenSvc := jwt.NewTokenService(tokenCfg)

	// Google
	googleCfg := oauth.GoogleOAuthConfig{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURL,
	}
	googleSvc := oauth.NewGoogleOAuthService(googleCfg)

	// State store for OAuth
	stateStore := oauth.NewStateStore()

	// Auth dependencies
	userRepo := postgres.NewUserRepository(dbPool, logr)
	authUC := usecase.NewAuthUseCase(userRepo, tokenSvc, googleSvc, val)
	authHandler := handler.NewAuthHandler(authUC, stateStore, cfg.FrontendURL)

	// Muzakki dependencies
	muzakkiRepo := postgres.NewMuzakkiRepository(dbPool, logr)
	muzakkiUC := usecase.NewMuzakkiUseCase(muzakkiRepo, val)
	muzakkiHandler := handler.NewMuzakkiHandler(muzakkiUC)

	// Asnaf dependencies
	asnafRepo := postgres.NewAsnafRepository(dbPool, logr)
	asnafUC := usecase.NewAsnafUseCase(asnafRepo, val)
	asnafHandler := handler.NewAsnafHandler(asnafUC)

	// Mustahiq dependencies
	mustahiqRepo := postgres.NewMustahiqRepository(dbPool, logr)
	mustahiqUC := usecase.NewMustahiqUseCase(mustahiqRepo, val)
	mustahiqHandler := handler.NewMustahiqHandler(mustahiqUC)

	// Program dependencies
	programRepo := postgres.NewProgramRepository(dbPool, logr)
	programUC := usecase.NewProgramUseCase(programRepo, val)
	programHandler := handler.NewProgramHandler(programUC)

	// DonationReceipt dependencies
	donationReceiptRepo := postgres.NewDonationReceiptRepository(dbPool, logr)
	donationReceiptUC := usecase.NewDonationReceiptUseCase(donationReceiptRepo, muzakkiRepo, val)
	donationReceiptHandler := handler.NewDonationReceiptHandler(donationReceiptUC)

	// Distribution dependencies
	distributionRepo := postgres.NewDistributionRepository(dbPool, logr)
	distributionUC := usecase.NewDistributionUseCase(distributionRepo, mustahiqRepo, val)
	distributionHandler := handler.NewDistributionHandler(distributionUC)

	// Report dependencies
	reportRepo := postgres.NewReportRepository(dbPool, logr)
	reportUC := usecase.NewReportUseCase(reportRepo, val)
	reportHandler := handler.NewReportHandler(reportUC)

	// User management dependencies
	userUC := usecase.NewUserUseCase(userRepo, val)
	userHandler := handler.NewUserHandler(userUC)

	// Middleware
	authMiddleware := middleware.NewAuthMiddleware(tokenSvc)

	router := gin.Default()

	// CORS middleware
	router.Use(func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		for _, o := range cfg.CORSAllowedOrigins {
			if o == origin {
				c.Header("Access-Control-Allow-Origin", origin)
			}
		}
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)

			auth.GET("/me", authMiddleware.RequireAuth(), authHandler.Me)

			auth.GET("/google/login", authHandler.GoogleLogin)
			auth.GET("/google/callback", authHandler.GoogleCallback)
			auth.POST("/google/mobile/login", authHandler.GoogleMobileLogin)
		}

		// Muzakki routes (protected)
		muzakki := v1.Group("/muzakki")
		muzakki.Use(authMiddleware.RequireAuth())
		{
			// GET - All authenticated users (viewer, staf, admin)
			muzakki.GET("", muzakkiHandler.FindAll)
			muzakki.GET("/:id", muzakkiHandler.FindByID)

			// POST, PUT - Staf and Admin only
			muzakki.POST("", authMiddleware.RequireStafOrAdmin(), muzakkiHandler.Create)
			muzakki.PUT("/:id", authMiddleware.RequireStafOrAdmin(), muzakkiHandler.Update)

			// DELETE - Admin only
			muzakki.DELETE("/:id", authMiddleware.RequireAdmin(), muzakkiHandler.Delete)
		}

		// Asnaf routes (protected)
		asnaf := v1.Group("/asnaf")
		asnaf.Use(authMiddleware.RequireAuth())
		{
			// GET - All authenticated users (viewer, staf, admin)
			asnaf.GET("", asnafHandler.FindAll)
			asnaf.GET("/:id", asnafHandler.FindByID)

			// POST, PUT, DELETE - Admin only
			asnaf.POST("", authMiddleware.RequireAdmin(), asnafHandler.Create)
			asnaf.PUT("/:id", authMiddleware.RequireAdmin(), asnafHandler.Update)
			asnaf.DELETE("/:id", authMiddleware.RequireAdmin(), asnafHandler.Delete)
		}

		// Mustahiq routes (protected)
		mustahiq := v1.Group("/mustahiq")
		mustahiq.Use(authMiddleware.RequireAuth())
		{
			// GET - All authenticated users (viewer, staf, admin)
			mustahiq.GET("", mustahiqHandler.FindAll)
			mustahiq.GET("/:id", mustahiqHandler.FindByID)

			// POST, PUT - Staf and Admin only
			mustahiq.POST("", authMiddleware.RequireStafOrAdmin(), mustahiqHandler.Create)
			mustahiq.PUT("/:id", authMiddleware.RequireStafOrAdmin(), mustahiqHandler.Update)

			// DELETE - Admin only
			mustahiq.DELETE("/:id", authMiddleware.RequireAdmin(), mustahiqHandler.Delete)
		}

		// Program routes (protected)
		programs := v1.Group("/programs")
		programs.Use(authMiddleware.RequireAuth())
		{
			// GET - All authenticated users (viewer, staf, admin)
			programs.GET("", programHandler.FindAll)
			programs.GET("/:id", programHandler.FindByID)

			// POST, PUT, DELETE - Admin only
			programs.POST("", authMiddleware.RequireAdmin(), programHandler.Create)
			programs.PUT("/:id", authMiddleware.RequireAdmin(), programHandler.Update)
			programs.DELETE("/:id", authMiddleware.RequireAdmin(), programHandler.Delete)
		}

		// DonationReceipt routes (protected)
		donationReceipts := v1.Group("/donation-receipts")
		donationReceipts.Use(authMiddleware.RequireAuth())
		{
			// GET - All authenticated users (viewer, staf, admin)
			donationReceipts.GET("", donationReceiptHandler.FindAll)
			donationReceipts.GET("/:id", donationReceiptHandler.FindByID)

			// POST, PUT - Staf and Admin only
			donationReceipts.POST("", authMiddleware.RequireStafOrAdmin(), donationReceiptHandler.Create)
			donationReceipts.PUT("/:id", authMiddleware.RequireStafOrAdmin(), donationReceiptHandler.Update)

			// DELETE - Admin only
			donationReceipts.DELETE("/:id", authMiddleware.RequireAdmin(), donationReceiptHandler.Delete)
		}

		// Distribution routes (protected)
		distributions := v1.Group("/distributions")
		distributions.Use(authMiddleware.RequireAuth())
		{
			// GET - All authenticated users (viewer, staf, admin)
			distributions.GET("", distributionHandler.FindAll)
			distributions.GET("/:id", distributionHandler.FindByID)

			// POST, PUT - Staf and Admin only
			distributions.POST("", authMiddleware.RequireStafOrAdmin(), distributionHandler.Create)
			distributions.PUT("/:id", authMiddleware.RequireStafOrAdmin(), distributionHandler.Update)

			// DELETE - Admin only
			distributions.DELETE("/:id", authMiddleware.RequireAdmin(), distributionHandler.Delete)
		}

		// Report routes (protected, read-only - All authenticated users)
		reports := v1.Group("/reports")
		reports.Use(authMiddleware.RequireAuth())
		{
			reports.GET("/income-summary", reportHandler.GetIncomeSummary)
			reports.GET("/distribution-summary", reportHandler.GetDistributionSummary)
			reports.GET("/fund-balance", reportHandler.GetFundBalance)
			reports.GET("/mustahiq-history/:mustahiq_id", reportHandler.GetMustahiqHistory)
		}

		// User Management routes (Admin only)
		users := v1.Group("/users")
		users.Use(authMiddleware.RequireAuth(), authMiddleware.RequireAdmin())
		{
			users.GET("", userHandler.FindAll)
			users.GET("/:id", userHandler.FindByID)
			users.PUT("/:id/role", userHandler.UpdateRole)
		}
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	srv := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: router,
	}

	go func() {
		logr.Infof("Server berjalan di :%s", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logr.Fatalf("Gagal ListenAndServe: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logr.Warn("Shutdown server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logr.Fatalf("Server Shutdown error: %v", err)
	}

	logr.Info("Server exited!")
}
