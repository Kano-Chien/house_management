package models

import "time"

type MealPlan struct {
	ID        int       `json:"id"`
	Date      time.Time `json:"date"`
	MealType  string    `json:"meal_type"` // Lunch, Dinner
	RecipeID  *int      `json:"recipe_id"`
	RecipeName string   `json:"recipe_name,omitempty"` // For display
}
