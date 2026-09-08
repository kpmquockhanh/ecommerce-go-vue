package models

import "time"

type Category struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	ProductCount int       `json:"product_count"`
	CreatedAt    time.Time `json:"created_at"`
}
