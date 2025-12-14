package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Lixing-Zhang/kart-challenge/backend-challenge/internal/config"
	"github.com/Lixing-Zhang/kart-challenge/backend-challenge/internal/coupon"
	"github.com/Lixing-Zhang/kart-challenge/backend-challenge/internal/handlers"
	"github.com/Lixing-Zhang/kart-challenge/backend-challenge/internal/middleware"
	"github.com/Lixing-Zhang/kart-challenge/backend-challenge/internal/repository"
	"github.com/Lixing-Zhang/kart-challenge/backend-challenge/internal/service"
	"github.com/Lixing-Zhang/kart-challenge/backend-challenge/pkg/logger"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	// Load configuration from environment
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize structured logger
	log := logger.New(cfg.LogLevel)
	slog.SetDefault(log)

	log.Info("starting food ordering api server",
		"port", cfg.Server.Port,
		"host", cfg.Server.Host,
		"log_level", cfg.LogLevel,
	)

	// Initialize coupon validator
	log.Info("loading coupon data...")
	couponValidator := coupon.NewValidator()
	couponURLs := []string{
		cfg.Coupon.File1URL,
		cfg.Coupon.File2URL,
		cfg.Coupon.File3URL,
	}
	
	ctx := context.Background()
	if err := couponValidator.LoadFromURLs(ctx, couponURLs); err != nil {
		log.Error("failed to load coupon data", "error", err)
		os.Exit(1)
	}
	
	stats := couponValidator.GetStats()
	log.Info("coupon data loaded successfully", 
		"total_files", stats["total_files"],
		"total_coupons", stats["total_coupons"],
	)

	// Initialize repositories
	productRepo := repository.NewInMemoryProductRepository()

	// Initialize services
	productService := service.NewProductService(productRepo)

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(log)
	productHandler := handlers.NewProductHandler(productService, log)
	couponHandler := handlers.NewCouponHandler(couponValidator)

	// Create router
	r := chi.NewRouter()

	// Apply middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.Logger(log))
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(60 * time.Second))

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "api_key"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Register health check endpoint
	r.Get("/health", healthHandler.ServeHTTP)

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Product endpoints
		r.Get("/product", productHandler.ListProducts)
		r.Get("/product/{productId}", productHandler.GetProduct)

		// Coupon endpoints
		r.Get("/coupon/{couponCode}", couponHandler.ValidateCoupon)
		r.Get("/coupon/stats", couponHandler.GetStats)

		// Order endpoints - to be implemented in next branch
	})

	// Create HTTP server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Info("server listening", "address", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Server.ShutdownTimeout)*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	log.Info("server stopped gracefully")
}
