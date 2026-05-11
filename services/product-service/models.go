package main

import "time"

type Product struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	Price       float64   `json:"price" gorm:"not null"`
	Stock       int       `json:"stock" gorm:"not null;default:0"`
	Category    string    `json:"category"`
	ImageURL    string    `json:"image_url"`
	SKU         string    `json:"sku" gorm:"unique"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UpdateStockRequest struct {
	Quantity int `json:"quantity"`
}

type ValidateBuyRequest struct {
	Quantity int `json:"quantity"`
}

type ValidateBuyResponse struct {
	Valid   bool    `json:"valid"`
	Message string  `json:"message,omitempty"`
	Stock   int     `json:"stock"`
	Price   float64 `json:"price"`
}
