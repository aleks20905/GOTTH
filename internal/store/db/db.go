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

	// Add Product, Cart, CartItem to migrations
	err = db.AutoMigrate(
		&store.User{},
		&store.Session{},
		&store.Schedule{},
		&store.Product{},  // NEW
		&store.Cart{},     // NEW
		&store.CartItem{}, // NEW
	)

	if err != nil {
		panic(err)
	}

	// Seed products if empty
	seedProducts(db)

	return db
}

func seedProducts(db *gorm.DB) {
	var count int64
	db.Model(&store.Product{}).Count(&count)

	if count == 0 {
		products := []store.Product{
			{Name: "Premium Ribeye Steak", Description: "Tender, juicy ribeye with perfect marbling", Price: 24.99, Image: "🥩", Weight: "400g", Category: "Beef", Stock: 50},
			{Name: "Bacon Strips", Description: "Crispy, smoky bacon strips", Price: 8.99, Image: "🥓", Weight: "250g", Category: "Pork", Stock: 100},
			{Name: "Chicken Wings", Description: "Fresh chicken wings, perfect for grilling", Price: 11.99, Image: "🍗", Weight: "500g", Category: "Poultry", Stock: 75},
			{Name: "Lamb Chops", Description: "Tender lamb chops with herbs", Price: 19.99, Image: "🍖", Weight: "350g", Category: "Lamb", Stock: 30},
			{Name: "Ground Beef", Description: "Premium ground beef, 80/20 blend", Price: 12.99, Image: "🥩", Weight: "500g", Category: "Beef", Stock: 60},
			{Name: "Pork Sausages", Description: "Homestyle pork sausages", Price: 9.99, Image: "🌭", Weight: "400g", Category: "Pork", Stock: 80},
		}
		db.Create(&products)
	}
}
