package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cor-events-scheduler/internal/config"
	"cor-events-scheduler/internal/handlers"
	"cor-events-scheduler/internal/handlers/middleware"
	"cor-events-scheduler/internal/infrastructure/db"
	"cor-events-scheduler/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	database, err := db.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize services
	schedulerService := services.NewSchedulerService(database)

	// Initialize router
	router := setupRouter(schedulerService)

	// Set Gin mode
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create HTTP server
	serverAddr := fmt.Sprintf("%s:%s", cfg.Server.Address, cfg.Server.Port)
	if cfg.Server.Address == "localhost" || cfg.Server.Address == "127.0.0.1" {
		serverAddr = fmt.Sprintf(":%s", cfg.Server.Port)
	}

	srv := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting server on %s", serverAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Shutdown gracefully
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

func setupRouter(schedulerService *services.SchedulerService) *gin.Engine {
	router := gin.Default()

	// Add middlewares
	router.Use(middleware.LoggingMiddleware())
	router.Use(middleware.CORSMiddleware())

	// Prometheus metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Health check
	router.GET("/health", handlers.HealthCheck)

	// API v1
	v1 := router.Group("/api/v1")
	{
		schedules := v1.Group("/schedules")
		{
			handler := handlers.NewSchedulerHandler(schedulerService)
			schedules.GET("/", handler.List)
			schedules.POST("/", handler.Create)
			schedules.PUT("/:id", handler.Update)
			schedules.GET("/:id", handler.Get)
			schedules.DELETE("/:id", handler.Delete)
			schedules.POST("/arrange", handler.ArrangeSchedule)
		}
	}

	return router
}
