package main

import (
	"goth/internal/store"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

func main() {
	logConfig := logger.Config{
		SlowThreshold: time.Second, // Increase slow query threshold
		LogLevel:      logger.Warn,
		Colorful:      true,
	}

	postgresDSN := os.Getenv("DATABASE_PUBLIC_URL")
	if postgresDSN == "" {
		log.Fatal("DATABASE_PUBLIC_URL environment variable is not set")
	}

	postgresDB, err := gorm.Open(postgres.Open(postgresDSN), &gorm.Config{
		Logger:                 logger.New(log.Default(), logConfig),
		PrepareStmt:            true, // Enable prepared statement cache
		SkipDefaultTransaction: true,
	})
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL database:", err)
	}

	// Connect to SQLite with optimized settings
	sqliteDB, err := gorm.Open(sqlite.Open("goth.db"), &gorm.Config{
		Logger:                 logger.New(log.Default(), logConfig),
		PrepareStmt:            true,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		log.Fatal("Failed to connect to SQLite database:", err)
	}

	// Create tables manually instead of using AutoMigrate
	err = sqliteDB.Exec(`
		CREATE TABLE IF NOT EXISTS schedules (
			id INTEGER PRIMARY KEY,
			course TEXT,
			spec TEXT,
			group_name TEXT,
			title TEXT,
			start DATETIME,
			end DATETIME,
			room TEXT,
			teacher TEXT,
			type TEXT,
			group_s TEXT,
			des TEXT
		)
	`).Error
	if err != nil {
		log.Fatal("Failed to create SQLite table:", err)
	}

	// Retrieve data from PostgreSQL in batches
	const batchSize = 1000
	var totalCount int64
	postgresDB.Model(&store.Schedule{}).Count(&totalCount)

	log.Printf("Total records to migrate: %d", totalCount)

	for offset := 0; offset < int(totalCount); offset += batchSize {
		var schedules []store.Schedule
		if err := postgresDB.Limit(batchSize).Offset(offset).Find(&schedules).Error; err != nil {
			log.Fatal("Failed to retrieve batch from PostgreSQL:", err)
		}

		//Insert batch into SQLite with upsert
		if len(schedules) > 0 {
			err := sqliteDB.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{
					"course", "spec", "group_name", "title",
					"start", "end", "room", "teacher",
					"type", "group_s", "des",
				}),
			}).CreateInBatches(&schedules, 100).Error

			if err != nil {
				log.Fatal("Failed to insert batch into SQLite:", err)
			}

			log.Printf("Migrated %d/%d records", offset+len(schedules), totalCount)
		}
	}

	log.Println("Successfully completed migration!")
}
