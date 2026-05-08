package models

import (
	"time"
)

type Product struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null" binding:"required"`
	Description string    `json:"description"`
	Price       float64   `json:"price" gorm:"not null" binding:"required,gt=0"`
	Stock       int       `json:"stock" gorm:"not null;default:0" binding:"min=0"`
	Category    string    `json:"category"`
	ImageURL    string    `json:"image_url"`
	SKU         string    `json:"sku" gorm:"unique" binding:"required"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
