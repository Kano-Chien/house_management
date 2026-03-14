package models

type Ingredient struct {
	ID                 int     `json:"id"`
	Name               string  `json:"name"`
	CurrentStock       int     `json:"current_stock"`
	MinStock           int     `json:"min_stock"`
	Price              int     `json:"price"`
	Category           string  `json:"category"`
	IsTracked          bool    `json:"is_tracked"`
	PlannedConsumption int     `json:"planned_consumption"` // Calculated, not stored directly
	EarliestExpiry     *string `json:"earliest_expiry"`     // Nullable: earliest expiry date across batches
}

type IngredientBatch struct {
	ID           int    `json:"id"`
	IngredientID int    `json:"ingredient_id"`
	Quantity     int    `json:"quantity"`
	ExpiryDate   string `json:"expiry_date"` // YYYY-MM-DD, may be empty
}
