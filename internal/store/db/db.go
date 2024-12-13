package db

import (
	"goth/internal/store"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite" // Sqlite driver based on CGO

	// "github.com/glebarez/sqlite" // Pure go SQLite driver, checkout https://github.com/glebarez/sqlite for details
	"gorm.io/gorm"
)

func open(dbName, dbUrl string) (*gorm.DB, error) {

	if os.Getenv("ENV") == "production" {
		return gorm.Open(postgres.Open(dbUrl), &gorm.Config{}) // TODO find a better place to handle common things

	}

	// make the temp directory if it doesn't exist
	err := os.MkdirAll("/tmp", 0755)
	if err != nil {
		return nil, err
	}

	return gorm.Open(sqlite.Open(dbName), &gorm.Config{})
}

func MustOpen(dbName, dbUrl string) *gorm.DB {

	if dbName == "" {
		dbName = "goth.db"
	}

	db, err := open(dbName, dbUrl)
	if err != nil {
		panic(err)
	}

	err = db.AutoMigrate(&store.User{}, &store.Session{}, &store.Schedule{})

	if err != nil {
		panic(err)
	}

	return db
}
