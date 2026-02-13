package models

import "time"

type Ingredient struct {
	ID                 int        `json:"id"`
	Name               string     `json:"name"`
	CurrentStock       float64    `json:"current_stock"`
	Unit               string     `json:"unit"`
	ExpiryDate         *time.Time `json:"expiry_date,omitempty"` // Pointer to allow null
	Price              float64    `json:"price"`
	Category           string     `json:"category"`
	PlannedConsumption float64    `json:"planned_consumption"` // Calculated, not stored directly
}
