package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Kano-Chien/house_management/backend/models"
)

type MealPlanHandler struct {
	DB *sql.DB
}

type ingredientInput struct {
	IngredientID int `json:"ingredient_id"`
	Quantity     int `json:"quantity"`
}

func (h *MealPlanHandler) GetMealPlan(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(`
		SELECT mp.id, mp.date, mp.meal_type, mp.recipe_id, r.name, COALESCE(mp.custom_name, ''), COALESCE(mp.is_cooked, FALSE)
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
		if err := rows.Scan(&mp.ID, &dateStr, &mp.MealType, &mp.RecipeID, &rName, &mp.CustomName, &mp.IsCooked); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Parse the date string from SQLite
		parsedDate, err := time.Parse("2006-01-02", dateStr[:10])
		if err == nil {
			mp.Date = parsedDate
		} else {
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
	type Request struct {
		Date        string           `json:"date"` // YYYY-MM-DD
		MealType    string           `json:"meal_type"`
		RecipeID    *int             `json:"recipe_id"`
		CustomName  string           `json:"custom_name"`
		Ingredients []ingredientInput `json:"ingredients"`
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

	tx, err := h.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	res, err := tx.Exec("INSERT INTO meal_plan (date, meal_type, recipe_id, custom_name) VALUES (?, ?, ?, ?)",
		date.Format("2006-01-02"), input.MealType, input.RecipeID, input.CustomName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id64, err := res.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id := int(id64)

	if input.Ingredients != nil {
		// User explicitly provided ingredients (even if empty — means they customized)
		for _, ing := range input.Ingredients {
			if _, err := tx.Exec("INSERT INTO meal_plan_ingredients (meal_plan_id, ingredient_id, quantity) VALUES (?, ?, ?)",
				id, ing.IngredientID, ing.Quantity); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	} else if input.RecipeID != nil {
		// No custom ingredients provided — copy from recipe as defaults
		_, err := tx.Exec(`
			INSERT INTO meal_plan_ingredients (meal_plan_id, ingredient_id, quantity)
			SELECT ?, ri.ingredient_id, ri.quantity
			FROM recipe_ingredients ri
			WHERE ri.recipe_id = ?
		`, id, *input.RecipeID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int{"id": id})
}

func (h *MealPlanHandler) UpdateMealPlan(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		ID          int              `json:"id"`
		CustomName  *string          `json:"custom_name"`
		Ingredients []ingredientInput `json:"ingredients"`
	}
	var input Request
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := h.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	if input.CustomName != nil {
		if _, err := tx.Exec("UPDATE meal_plan SET custom_name = ? WHERE id = ?", *input.CustomName, input.ID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if input.Ingredients != nil {
		if _, err := tx.Exec("DELETE FROM meal_plan_ingredients WHERE meal_plan_id = ?", input.ID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, ing := range input.Ingredients {
			if _, err := tx.Exec("INSERT INTO meal_plan_ingredients (meal_plan_id, ingredient_id, quantity) VALUES (?, ?, ?)",
				input.ID, ing.IngredientID, ing.Quantity); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
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

	// 3. Decrement Inventory — use meal_plan_ingredients (customized per meal)
	type ingredientUpdate struct {
		ID   int
		Name string
		Qty  int
	}

	scanIngredientUpdates := func(rows *sql.Rows) ([]ingredientUpdate, []string, error) {
		var upd []ingredientUpdate
		var missing []string
		for rows.Next() {
			var u ingredientUpdate
			var currentStock int
			if err := rows.Scan(&u.ID, &u.Name, &u.Qty, &currentStock); err != nil {
				rows.Close()
				return nil, nil, err
			}
			if currentStock < u.Qty {
				missing = append(missing, fmt.Sprintf("%s (Need: %d, Have: %d)", u.Name, u.Qty, currentStock))
			}
			upd = append(upd, u)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return nil, nil, err
		}
		return upd, missing, nil
	}

	// Try meal_plan_ingredients first; fall back to recipe_ingredients for backward compat
	rows, err := tx.Query(`
		SELECT mpi.ingredient_id, i.name, mpi.quantity, i.current_stock
		FROM meal_plan_ingredients mpi
		JOIN ingredients i ON mpi.ingredient_id = i.id
		WHERE mpi.meal_plan_id = ? AND i.is_tracked = TRUE
	`, req.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	updates, missingIngredients, err := scanIngredientUpdates(rows)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Fallback: if no meal_plan_ingredients, backfill from recipe_ingredients
	if len(updates) == 0 {
		// Backfill meal_plan_ingredients from recipe so future cooks use the right source
		_, err = tx.Exec(`
			INSERT INTO meal_plan_ingredients (meal_plan_id, ingredient_id, quantity)
			SELECT ?, ri.ingredient_id, ri.quantity
			FROM recipe_ingredients ri
			WHERE ri.recipe_id = ?
		`, req.ID, recipeID.Int64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		rows, err = tx.Query(`
			SELECT mpi.ingredient_id, i.name, mpi.quantity, i.current_stock
			FROM meal_plan_ingredients mpi
			JOIN ingredients i ON mpi.ingredient_id = i.id
			WHERE mpi.meal_plan_id = ? AND i.is_tracked = TRUE
		`, req.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		updates, missingIngredients, err = scanIngredientUpdates(rows)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if len(missingIngredients) > 0 {
		http.Error(w, fmt.Sprintf("Not enough stock: %s", strings.Join(missingIngredients, ", ")), http.StatusConflict)
		return
	}

	// Apply updates - FEFO: consume from batches ordered by earliest expiry first (nulls last)
	for _, u := range updates {
		remaining := u.Qty

		batchRows, err := tx.Query(`
			SELECT id, quantity FROM ingredient_batches
			WHERE ingredient_id = ?
			ORDER BY CASE WHEN expiry_date IS NULL OR expiry_date = '' THEN 1 ELSE 0 END,
			         expiry_date ASC, id ASC
		`, u.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		type batchItem struct {
			ID  int
			Qty int
		}
		var batches []batchItem
		for batchRows.Next() {
			var b batchItem
			if err := batchRows.Scan(&b.ID, &b.Qty); err != nil {
				batchRows.Close()
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			batches = append(batches, b)
		}
		batchRows.Close()

		for _, b := range batches {
			if remaining <= 0 {
				break
			}
			consume := b.Qty
			if consume > remaining {
				consume = remaining
			}
			remaining -= consume
			newQty := b.Qty - consume
			if newQty == 0 {
				if _, err := tx.Exec("DELETE FROM ingredient_batches WHERE id = ?", b.ID); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			} else {
				if _, err := tx.Exec("UPDATE ingredient_batches SET quantity = ? WHERE id = ?", newQty, b.ID); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}

		// Recompute current_stock as sum of remaining batches (consistent with UpdateBatch/DeleteBatch)
		if _, err := tx.Exec(
			"UPDATE ingredients SET current_stock = (SELECT COALESCE(SUM(quantity), 0) FROM ingredient_batches WHERE ingredient_id = ?) WHERE id = ?",
			u.ID, u.ID,
		); err != nil {
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

func (h *MealPlanHandler) GetMealPlanIngredients(w http.ResponseWriter, r *http.Request) {
	mealPlanIDStr := r.URL.Query().Get("meal_plan_id")
	if mealPlanIDStr == "" {
		http.Error(w, "meal_plan_id required", http.StatusBadRequest)
		return
	}
	mealPlanID, err := strconv.Atoi(mealPlanIDStr)
	if err != nil {
		http.Error(w, "meal_plan_id must be a valid integer", http.StatusBadRequest)
		return
	}

	rows, err := h.DB.Query(`
		SELECT mpi.ingredient_id, i.name, mpi.quantity
		FROM meal_plan_ingredients mpi
		JOIN ingredients i ON mpi.ingredient_id = i.id
		WHERE mpi.meal_plan_id = ?
	`, mealPlanID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type IngredientDetail struct {
		IngredientID int    `json:"ingredient_id"`
		Name         string `json:"name"`
		Quantity     int    `json:"quantity"`
	}

	var ingredients []IngredientDetail
	for rows.Next() {
		var ing IngredientDetail
		if err := rows.Scan(&ing.IngredientID, &ing.Name, &ing.Quantity); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ingredients = append(ingredients, ing)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ingredients)
}
