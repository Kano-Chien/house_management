package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Kano-Chien/house_management/backend/models"
)

type MealPlanHandler struct {
	DB *sql.DB
}

func (h *MealPlanHandler) GetMealPlan(w http.ResponseWriter, r *http.Request) {
	// Optional: Filter by date range query params ?start=...&end=...
	rows, err := h.DB.Query(`
		SELECT mp.id, mp.date, mp.meal_type, mp.recipe_id, r.name, COALESCE(mp.is_cooked, FALSE)
		FROM meal_plan mp 
		LEFT JOIN recipes r ON mp.recipe_id = r.id
		ORDER BY mp.date, mp.meal_type
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var plan []models.MealPlan
	for rows.Next() {
		var mp models.MealPlan
		var dateStr string
		var rName sql.NullString
		if err := rows.Scan(&mp.ID, &dateStr, &mp.MealType, &mp.RecipeID, &rName, &mp.IsCooked); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Parse the date string from SQLite
		parsedDate, err := time.Parse("2006-01-02", dateStr[:10]) // Take first 10 chars just in case
		if err == nil {
			mp.Date = parsedDate
		} else {
			// Fallback or log error, but keep going?
			// Try RFC3339 just in case
			if parsedDate, err := time.Parse(time.RFC3339, dateStr); err == nil {
				mp.Date = parsedDate
			}
		}

		if rName.Valid {
			mp.RecipeName = rName.String
		}
		plan = append(plan, mp)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plan)
}

func (h *MealPlanHandler) ScheduleMeal(w http.ResponseWriter, r *http.Request) {
	// Fix date parsing if JSON sends string, but let's assume standard ISO8601 handled by Go's JSON parser to time.Time if format matches
	// Or use a custom struct for decoding
	type Request struct {
		Date     string `json:"date"` // YYYY-MM-DD
		MealType string `json:"meal_type"`
		RecipeID *int   `json:"recipe_id"`
	}
	var input Request
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	date, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	var id int
	res, err := h.DB.Exec("INSERT INTO meal_plan (date, meal_type, recipe_id) VALUES (?, ?, ?)", date.Format("2006-01-02"), input.MealType, input.RecipeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id64, err := res.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id = int(id64)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int{"id": id})
}

func (h *MealPlanHandler) DeleteMealPlan(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err := h.DB.Exec("DELETE FROM meal_plan WHERE id = ?", req.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

func (h *MealPlanHandler) CookMeal(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := h.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// 1. Check current status and get recipe ID
	var recipeID sql.NullInt64
	var isCooked bool
	err = tx.QueryRow("SELECT recipe_id, COALESCE(is_cooked, FALSE) FROM meal_plan WHERE id = ?", req.ID).Scan(&recipeID, &isCooked)
	if err == sql.ErrNoRows {
		http.Error(w, "Meal plan not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if isCooked {
		http.Error(w, "Meal already cooked", http.StatusConflict)
		return
	}

	if !recipeID.Valid {
		http.Error(w, "No recipe associated with this meal", http.StatusBadRequest)
		return
	}

	// 2. Mark as cooked
	_, err = tx.Exec("UPDATE meal_plan SET is_cooked = TRUE WHERE id = ?", req.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Decrement Inventory
	// Buffer the ingredient updates because we cannot interleave Exec with open Query rows on the same tx
	type ingredientUpdate struct {
		ID   int
		Name string
		Qty  float64
	}
	var updates []ingredientUpdate
	var missingIngredients []string

	rows, err := tx.Query(`
		SELECT ri.ingredient_id, i.name, ri.quantity, i.current_stock
		FROM recipe_ingredients ri
		JOIN ingredients i ON ri.ingredient_id = i.id
		WHERE ri.recipe_id = ? AND i.is_tracked = TRUE
	`, recipeID.Int64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for rows.Next() {
		var u ingredientUpdate
		var currentStock float64
		if err := rows.Scan(&u.ID, &u.Name, &u.Qty, &currentStock); err != nil {
			rows.Close()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if currentStock < u.Qty {
			missingIngredients = append(missingIngredients, fmt.Sprintf("%s (Need: %.2f, Have: %.2f)", u.Name, u.Qty, currentStock))
		}
		updates = append(updates, u)
	}
	rows.Close()

	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(missingIngredients) > 0 {
		http.Error(w, fmt.Sprintf("Not enough stock: %s", strings.Join(missingIngredients, ", ")), http.StatusConflict)
		return
	}

	// Apply updates
	for _, u := range updates {
		_, err = tx.Exec("UPDATE ingredients SET current_stock = current_stock - ? WHERE id = ?", u.Qty, u.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "cooked"})
}
