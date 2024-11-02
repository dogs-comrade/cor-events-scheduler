// cmd/app/main.go
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cor-events-scheduler/internal/config"
	"cor-events-scheduler/internal/domain/repositories"
	"cor-events-scheduler/internal/handlers"
	"cor-events-scheduler/internal/handlers/middleware"
	"cor-events-scheduler/internal/infrastructure/db"
	"cor-events-scheduler/internal/metrics"
	"cor-events-scheduler/internal/services"
	"cor-events-scheduler/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger := utils.InitLogger()
	defer logger.Sync()

	// Initialize metrics
	metrics.InitMetrics()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	// Initialize database
	database, err := db.NewDatabase(cfg)
	if err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}

	// Initialize repositories
	scheduleRepo := repositories.NewScheduleRepository(database)
	eventRepo := repositories.NewEventRepository(database)

	// Initialize services
	analysisService := services.NewAnalysisService(cfg)
	schedulerService := services.NewSchedulerService(
		scheduleRepo,
		eventRepo,
		analysisService,
		logger,
		database,
	)

	// Initialize router
	router := setupRouter(schedulerService, logger)

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
		logger.Info("Starting server", zap.String("address", serverAddr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Shutdown gracefully
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited properly")
}

func setupRouter(schedulerService *services.SchedulerService, logger *zap.Logger) *gin.Engine {
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(middleware.NewLoggingMiddleware(logger))
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.NewMetricsMiddleware())

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	router.GET("/health", handlers.HealthCheck)

	v1 := router.Group("/api/v1")
	{
		schedules := v1.Group("/schedules")
		{
			handler := handlers.NewSchedulerHandler(schedulerService, logger)
			schedules.POST("/", handler.CreateSchedule)
			schedules.GET("/", handler.ListSchedules)
			schedules.GET("/:id", handler.GetSchedule)
			schedules.PUT("/:id", handler.UpdateSchedule)
			schedules.DELETE("/:id", handler.DeleteSchedule)
			schedules.POST("/analyze", handler.AnalyzeSchedule)
			schedules.POST("/optimize", handler.OptimizeSchedule)
		}

		// События пока закомментируем, так как нет имплементации
		/*
		   events := v1.Group("/events")
		   {
		       handler := handlers.NewEventHandler(schedulerService, logger)
		       events.POST("/:id/schedules", handler.CreateEventSchedule)
		       events.GET("/:id/schedules", handler.GetEventSchedules)
		       events.GET("/:id/analysis", handler.GetEventAnalysis)
		   }
		*/
	}

	return router
}
