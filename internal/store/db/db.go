package db

import (
	"goth/internal/store"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"

	"gorm.io/gorm"
)

func open(dbName, dbUrl string) (*gorm.DB, error) {

	if os.Getenv("ENV") == "production" {
		return gorm.Open(postgres.Open(dbUrl), &gorm.Config{}) // TODO find a better place to handle common things

	}

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

	err = db.AutoMigrate(
		&store.User{},
		&store.Session{},
		&store.Device{},
		&store.DeviceGroup{},
		&store.FirewallRule{},
		&store.UpdateTask{},
		&store.LogEntry{},
	)

	if err != nil {
		panic(err)
	}

	return db
}
