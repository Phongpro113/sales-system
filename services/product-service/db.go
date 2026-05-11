package main

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func initDB() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=postgres user=postgres password=postgres dbname=product_db port=5432 sslmode=disable"
	}

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err := db.AutoMigrate(&Product{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	var count int64
	db.Model(&Product{}).Count(&count)
	if count == 0 {
		seedProducts()
	}
}

func seedProducts() {
	products := []Product{
		{
			Name:        "MacBook Pro 14",
			Description: "Apple M3 Pro chip, 16GB RAM, 512GB SSD",
			Price:       1999.99,
			Stock:       25,
			Category:    "Electronics",
			ImageURL:    "https://images.unsplash.com/photo-1517336714731-489689fd1ca8",
			SKU:         "MBP-14-M3-16-512",
		},
		{
			Name:        "Sony WH-1000XM5",
			Description: "Wireless Noise Canceling Headphones",
			Price:       399.99,
			Stock:       50,
			Category:    "Electronics",
			ImageURL:    "https://images.unsplash.com/photo-1618366712010-f4ae9c647dcb",
			SKU:         "SNY-XM5-BLK",
		},
		{
			Name:        "Nike Air Max",
			Description: "Running Shoes, breathable mesh upper",
			Price:       129.99,
			Stock:       100,
			Category:    "Clothing",
			ImageURL:    "https://images.unsplash.com/photo-1542291026-7eec264c27ff",
			SKU:         "NKE-AM-BLK",
		},
		{
			Name:        "The Go Programming Language",
			Description: "By Alan A. A. Donovan & Brian W. Kernighan",
			Price:       49.99,
			Stock:       75,
			Category:    "Books",
			ImageURL:    "https://images.unsplash.com/photo-1589998059171-988d887df646",
			SKU:         "BOOK-GO-001",
		},
		{
			Name:        "Mechanical Keyboard",
			Description: "RGB Backlit Mechanical Gaming Keyboard",
			Price:       89.99,
			Stock:       150,
			Category:    "Electronics",
			ImageURL:    "https://images.unsplash.com/photo-1595225476474-87563907a212",
			SKU:         "KEY-MECH-RGB",
		},
		{
			Name:        "Coffee Mug",
			Description: "Ceramic Coffee Mug, 15oz",
			Price:       14.99,
			Stock:       200,
			Category:    "Home",
			ImageURL:    "https://images.unsplash.com/photo-1514228742587-6b1558fcca3d",
			SKU:         "MUG-CER-BLK",
		},
	}

	for _, product := range products {
		db.Create(&product)
	}
	log.Println("Seeded database with sample products")
}
