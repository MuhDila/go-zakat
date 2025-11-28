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
	"go-zakat/pkg/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"go-zakat/docs"

	"go-zakat/internal/delivery/http/handler"
	"go-zakat/internal/delivery/http/middleware"
	domainValidator "go-zakat/internal/delivery/http/validator"
	"go-zakat/internal/infrastructure/jwt"
	"go-zakat/internal/infrastructure/oauth"
	"go-zakat/internal/repository/postgres"
	"go-zakat/internal/usecase"

	"go-zakat/pkg/config"
	"go-zakat/pkg/database"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()
	logr := logger.New() // <-- logger init

	// Swagger info
	docs.SwaggerInfo.Title = "Auth API"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:" + cfg.AppPort

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
	authHandler := handler.NewAuthHandler(authUC, stateStore)

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
			muzakki.GET("", muzakkiHandler.FindAll)
			muzakki.GET("/:id", muzakkiHandler.FindByID)
			muzakki.POST("", muzakkiHandler.Create)
			muzakki.PUT("/:id", muzakkiHandler.Update)
			muzakki.DELETE("/:id", muzakkiHandler.Delete)
		}

		// Asnaf routes (protected)
		asnaf := v1.Group("/asnaf")
		asnaf.Use(authMiddleware.RequireAuth())
		{
			asnaf.GET("", asnafHandler.FindAll)
			asnaf.GET("/:id", asnafHandler.FindByID)
			asnaf.POST("", asnafHandler.Create)
			asnaf.PUT("/:id", asnafHandler.Update)
			asnaf.DELETE("/:id", asnafHandler.Delete)
		}

		// Mustahiq routes (protected)
		mustahiq := v1.Group("/mustahiq")
		mustahiq.Use(authMiddleware.RequireAuth())
		{
			mustahiq.GET("", mustahiqHandler.FindAll)
			mustahiq.GET("/:id", mustahiqHandler.FindByID)
			mustahiq.POST("", mustahiqHandler.Create)
			mustahiq.PUT("/:id", mustahiqHandler.Update)
			mustahiq.DELETE("/:id", mustahiqHandler.Delete)
		}

		// Program routes (protected)
		programs := v1.Group("/programs")
		programs.Use(authMiddleware.RequireAuth())
		{
			programs.GET("", programHandler.FindAll)
			programs.GET("/:id", programHandler.FindByID)
			programs.POST("", programHandler.Create)
			programs.PUT("/:id", programHandler.Update)
			programs.DELETE("/:id", programHandler.Delete)
		}

		// DonationReceipt routes (protected)
		donationReceipts := v1.Group("/donation-receipts")
		donationReceipts.Use(authMiddleware.RequireAuth())
		{
			donationReceipts.GET("", donationReceiptHandler.FindAll)
			donationReceipts.GET("/:id", donationReceiptHandler.FindByID)
			donationReceipts.POST("", donationReceiptHandler.Create)
			donationReceipts.PUT("/:id", donationReceiptHandler.Update)
			donationReceipts.DELETE("/:id", donationReceiptHandler.Delete)
		}

		// Distribution routes (protected)
		distributions := v1.Group("/distributions")
		distributions.Use(authMiddleware.RequireAuth())
		{
			distributions.GET("", distributionHandler.FindAll)
			distributions.GET("/:id", distributionHandler.FindByID)
			distributions.POST("", distributionHandler.Create)
			distributions.PUT("/:id", distributionHandler.Update)
			distributions.DELETE("/:id", distributionHandler.Delete)
		}

		// Report routes (protected, read-only)
		reports := v1.Group("/reports")
		reports.Use(authMiddleware.RequireAuth())
		{
			reports.GET("/income-summary", reportHandler.GetIncomeSummary)
			reports.GET("/distribution-summary", reportHandler.GetDistributionSummary)
			reports.GET("/fund-balance", reportHandler.GetFundBalance)
			reports.GET("/mustahiq-history/:mustahiq_id", reportHandler.GetMustahiqHistory)
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
