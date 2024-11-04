// internal/infrastructure/db/database.go
package db

import (
	"cor-events-scheduler/internal/config"
	"cor-events-scheduler/internal/domain/models"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Обновленная миграция только для необходимых моделей
	if err := db.AutoMigrate(
		&models.Schedule{},
		&models.Block{},
		&models.BlockItem{},
		&models.ScheduleVersion{},
	); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database connection established and migrations completed")
	return db, nil
}
