package models

import "time"

type MealPlan struct {
	ID         int       `json:"id"`
	Date       time.Time `json:"date"`
	MealType   string    `json:"meal_type"` // Breakfast, Lunch, Dinner
	RecipeID   *int      `json:"recipe_id"`
	RecipeName string    `json:"recipe_name,omitempty"` // For display (from JOIN)
	CustomName string    `json:"custom_name,omitempty"` // For meals without a recipe
	IsCooked   bool      `json:"is_cooked"`
}
