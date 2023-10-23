package main

import (
	"time"
)

type Product struct {
	ID               int64     `json:"id"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	Images           []string  `json:"images"`
	Price            float64   `json:"price"`
	UserID           int64     `json:"user_id"`
	CompressedImages []string  `json:"compressed_images"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Mobile    string    `json:"mobile"`
	Latitude  string    `json:"latitude"`
	Longitude string    `json:"longitude"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
