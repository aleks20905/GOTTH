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
)

func main() {
	postgresDSN := os.Getenv("DATABASE_PUBLIC_URL")
	if postgresDSN == "" {
		log.Fatal("DATABASE_PUBLIC_URL environment variable is not set")
	}

	postgresDB, err := gorm.Open(postgres.Open(postgresDSN), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL database:", err)
	}

	sqliteDB, err := gorm.Open(sqlite.Open("goth.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to SQLite database:", err)
	}

	err = sqliteDB.AutoMigrate(&store.Schedule{})
	if err != nil {
		log.Fatal("Failed to migrate SQLite schema:", err)
	}

	var schedules []store.Schedule
	if err := postgresDB.Find(&schedules).Error; err != nil {
		log.Fatal("Failed to retrieve data from PostgreSQL:", err)
	}

	log.Printf("Found %d schedules in PostgreSQL", len(schedules))

	for i, schedule := range schedules {
		schedules[i].Start = convertToUTC(schedule.Start)
		schedules[i].End = convertToUTC(schedule.End)

		log.Printf("Converted schedule %d to UTC: %+v\n", i+1, schedules[i])
	}

	if len(schedules) > 0 {
		err := sqliteDB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"course", "spec", "group_name", "title", "start", "end", "room", "teacher", "type", "group_s", "des"}),
		}).CreateInBatches(&schedules, 100).Error // Insert 100 rows at a time
		if err != nil {
			log.Fatal("Failed to insert or update data in SQLite:", err)
		}
		log.Println("Successfully migrated data to SQLite with upserts!")
	}
}

func convertToUTC(t time.Time) time.Time {
	return t.UTC() // Ensure the time is converted to UTC
}
