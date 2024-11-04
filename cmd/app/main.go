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

	"cor-events-scheduler/docs"
	_ "cor-events-scheduler/docs"
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
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

// @title Event Scheduler API
// @version 1.0
// @description Service for managing event schedules with risk analysis and optimization
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host scheduler.xilonen.ru
// @BasePath /api/v1
func main() {
	logger := utils.InitLogger()
	defer logger.Sync()

	metrics.InitMetrics()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	database, err := db.NewDatabase(cfg)
	if err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}

	docs.SwaggerInfo.Title = "Event Scheduler API"
	docs.SwaggerInfo.Description = "Service for managing event schedules with risk analysis and optimization"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:8282"
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	scheduleRepo := repositories.NewScheduleRepository(database)
	versionRepo := repositories.NewVersionRepository(database)

	versionService := services.NewVersionService(versionRepo, scheduleRepo, logger)

	schedulerService := services.NewSchedulerService(
		scheduleRepo,
		versionRepo,
		logger,
	)

	router := setupRouter(schedulerService, versionService, logger) // Добавляем logger

	docs.SwaggerInfo.Title = "Event Scheduler API"
	docs.SwaggerInfo.Description = "Service for managing event schedules with risk analysis and optimization"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:8282"
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

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

func setupRouter(
	schedulerService *services.SchedulerService,
	versionService *services.VersionService,
	logger *zap.Logger,
) *gin.Engine {
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(middleware.NewLoggingMiddleware(logger))
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.NewMetricsMiddleware())

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	formatterService := services.NewFormatterService(schedulerService, logger)
	formatterHandler := handlers.NewFormatterHandler(formatterService, logger)

	url := ginSwagger.URL("http://localhost:8282/swagger/doc.json")
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

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
			schedules.GET("/:id/public", formatterHandler.GetPublicSchedule)
		}
	}
	return router
}
