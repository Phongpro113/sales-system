package main

import "time"

type Order struct {
	ID              uint        `json:"id" gorm:"primaryKey"`
	OrderNumber     string      `json:"order_number" gorm:"unique;not null"`
	UserID          uint        `json:"user_id" gorm:"not null;index"`
	Total           float64     `json:"total"`
	Status          string      `json:"status" gorm:"default:pending"`
	PaymentMethod   string      `json:"payment_method"`
	ShippingAddress string      `json:"shipping_address"`
	Notes           string      `json:"notes"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
	Items           []OrderItem `json:"items" gorm:"foreignKey:OrderID"`
}

type OrderItem struct {
	ID          uint    `json:"id" gorm:"primaryKey"`
	OrderID     uint    `json:"order_id" gorm:"index"`
	ProductID   uint    `json:"product_id" gorm:"not null"`
	ProductName string  `json:"product_name" gorm:"-"`
	Quantity    int     `json:"quantity" gorm:"not null"`
	Price       float64 `json:"price" gorm:"not null"`
	Subtotal    float64 `json:"subtotal" gorm:"-"`
}

type Product struct {
	ID    uint    `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

type CreateOrderRequest struct {
	Items []struct {
		ProductID uint `json:"product_id"`
		Quantity  int  `json:"quantity"`
	} `json:"items"`
	PaymentMethod   string `json:"payment_method"`
	ShippingAddress string `json:"shipping_address"`
	Notes           string `json:"notes"`
}
