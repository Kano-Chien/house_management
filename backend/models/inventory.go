package models

type Ingredient struct {
	ID                 int     `json:"id"`
	Name               string  `json:"name"`
	CurrentStock       int     `json:"current_stock"`
	MinStock           int     `json:"min_stock"`
	Price              float64 `json:"price"`
	Category           string  `json:"category"`
	IsTracked          bool    `json:"is_tracked"`
	PlannedConsumption int     `json:"planned_consumption"` // Calculated, not stored directly
}
