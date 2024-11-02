// internal/infrastructure/db/database.go
package db

import (
	"fmt"
	"log"

	"cor-events-scheduler/internal/config"
	"cor-events-scheduler/internal/domain/models"

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

	// Создаем базовые таблицы
	if err := db.Exec(`
        -- Таблица для оборудования
        CREATE TABLE IF NOT EXISTS equipment (
            id BIGSERIAL PRIMARY KEY,
            name TEXT NOT NULL,
            type TEXT,
            setup_time INTEGER,
            complexity_score DECIMAL,
            created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            deleted_at TIMESTAMPTZ
        );

        -- Таблица для участников
        CREATE TABLE IF NOT EXISTS participants (
            id BIGSERIAL PRIMARY KEY,
            name TEXT NOT NULL,
            role TEXT,
            block_item_id BIGINT,
            requirements TEXT,
            created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            deleted_at TIMESTAMPTZ
        );

        -- Таблица для расписаний
        CREATE TABLE IF NOT EXISTS schedules (
            id BIGSERIAL PRIMARY KEY,
            event_id BIGINT,
            name TEXT NOT NULL,
            description TEXT,
            start_date TIMESTAMPTZ NOT NULL,
            end_date TIMESTAMPTZ NOT NULL,
            risk_score DECIMAL,
            buffer_time INTEGER,
            total_duration INTEGER,
            created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            deleted_at TIMESTAMPTZ
        );

        -- Таблица для блоков
        CREATE TABLE IF NOT EXISTS blocks (
            id BIGSERIAL PRIMARY KEY,
            schedule_id BIGINT REFERENCES schedules(id),
            name TEXT NOT NULL,
            type TEXT,
            start_time TIMESTAMPTZ,
            duration INTEGER,
            tech_break_duration INTEGER,
            complexity DECIMAL,
            max_participants INTEGER,
            required_staff INTEGER,
            location TEXT,
            dependencies INTEGER[],
            risk_factors JSONB,
            "order" INTEGER,
            created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            deleted_at TIMESTAMPTZ
        );

        -- Таблица для элементов блока
        CREATE TABLE IF NOT EXISTS block_items (
            id BIGSERIAL PRIMARY KEY,
            block_id BIGINT REFERENCES blocks(id),
            type TEXT,
            name TEXT NOT NULL,
            description TEXT,
            duration INTEGER,
            requirements TEXT,
            "order" INTEGER,
            created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            deleted_at TIMESTAMPTZ
        );

        -- Промежуточные таблицы для many-to-many связей
        CREATE TABLE IF NOT EXISTS block_equipment (
            block_id BIGINT REFERENCES blocks(id),
            equipment_id BIGINT REFERENCES equipment(id),
            created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            PRIMARY KEY (block_id, equipment_id)
        );

        CREATE TABLE IF NOT EXISTS block_item_equipment (
            block_item_id BIGINT REFERENCES block_items(id),
            equipment_id BIGINT REFERENCES equipment(id),
            created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            PRIMARY KEY (block_item_id, equipment_id)
        );

        -- Индексы
        CREATE INDEX IF NOT EXISTS idx_schedules_event_id ON schedules(event_id);
        CREATE INDEX IF NOT EXISTS idx_blocks_schedule_id ON blocks(schedule_id);
        CREATE INDEX IF NOT EXISTS idx_block_items_block_id ON block_items(block_id);
        CREATE INDEX IF NOT EXISTS idx_equipment_type ON equipment(type);
        CREATE INDEX IF NOT EXISTS idx_blocks_type ON blocks(type);
        CREATE INDEX IF NOT EXISTS idx_block_items_type ON block_items(type);
    `).Error; err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	// Запускаем GORM AutoMigrate для обновления структуры таблиц
	if err := db.AutoMigrate(
		&models.Equipment{},
		&models.Participant{},
		&models.Schedule{},
		&models.Block{},
		&models.BlockItem{},
	); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database connection established and migrations completed")
	return db, nil
}
