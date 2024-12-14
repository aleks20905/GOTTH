package main

import (
	"goth/internal/store"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func main() {
	// 1. Connect to the SQLite database
	sqliteDB, err := gorm.Open(sqlite.Open("goth.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to SQLite database:", err)
	}

	// 2. Connect to the PostgreSQL database
	postgresDSN := "postgresql://postgres:XrhFPTfkZumrlGFpOjwXKSneLcBpopRw@junction.proxy.rlwy.net:17865/railway"
	postgresDB, err := gorm.Open(postgres.Open(postgresDSN), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL database:", err)
	}

	// Ensure PostgreSQL database has the same schema by migrating
	err = postgresDB.AutoMigrate(&store.Schedule{})
	if err != nil {
		log.Fatal("Failed to migrate PostgreSQL schema:", err)
	}

	// 3. Retrieve all records from the SQLite database
	var schedules []store.Schedule
	if err := sqliteDB.Find(&schedules).Error; err != nil {
		log.Fatal("Failed to retrieve data from SQLite:", err)
	}

	log.Printf("Found %d schedules in SQLite", len(schedules))

	if len(schedules) > 0 {
		err := postgresDB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"course", "spec", "group_name", "title", "start", "end", "room", "teacher", "type", "group_s", "des"}),
		}).CreateInBatches(&schedules, 100).Error // Insert 100 rows at a time
		if err != nil {
			log.Fatal("Failed to insert or update data in PostgreSQL:", err)
		}
		log.Println("Successfully migrated data to PostgreSQL with upserts!")
	}
	// 4. Insert the data into PostgreSQL, update on conflict
}
