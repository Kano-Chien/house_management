package models

type Recipe struct {
	ID           int                `json:"id"`
	Name         string             `json:"name"`
	Instructions string             `json:"instructions"`
	Notes        string             `json:"notes"`
	Ingredients  []RecipeIngredient `json:"ingredients,omitempty"`
}

type RecipeIngredient struct {
	IngredientID int     `json:"ingredient_id"`
	Name         string  `json:"name,omitempty"` // For display
	Quantity     float64 `json:"quantity"`
	Unit         string  `json:"unit,omitempty"` // For display
}
