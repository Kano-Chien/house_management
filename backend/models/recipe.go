package models

type Recipe struct {
	ID          int                `json:"id"`
	Name        string             `json:"name"`
	Notes       string             `json:"notes"`
	Ingredients []RecipeIngredient `json:"ingredients,omitempty"`
}

type RecipeIngredient struct {
	IngredientID int    `json:"ingredient_id"`
	Name         string `json:"name,omitempty"` // For display
	Quantity     int    `json:"quantity"`
}
