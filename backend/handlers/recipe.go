package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/Kano-Chien/house_management/backend/models"
)

type RecipeHandler struct {
	DB *sql.DB
}

func (h *RecipeHandler) GetRecipes(w http.ResponseWriter, r *http.Request) {
	// Simple fetch for listing. Detailed fetch with ingredients could be separate or joined.
	// For MVP, letting's just fetch basic info.
	rows, err := h.DB.Query("SELECT id, name, instructions, COALESCE(notes, '') as notes FROM recipes")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var recipes []models.Recipe
	for rows.Next() {
		var r models.Recipe
		if err := rows.Scan(&r.ID, &r.Name, &r.Instructions, &r.Notes); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		recipes = append(recipes, r)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recipes)
}

func (h *RecipeHandler) CreateRecipe(w http.ResponseWriter, r *http.Request) {
	var req models.Recipe
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

	var recipeID int
	err = tx.QueryRow("INSERT INTO recipes (name, instructions) VALUES ($1, $2) RETURNING id", req.Name, req.Instructions).Scan(&recipeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Insert ingredients if provided
	if len(req.Ingredients) > 0 {
		stmt, err := tx.Prepare("INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity) VALUES ($1, $2, $3)")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer stmt.Close()

		for _, ing := range req.Ingredients {
			if _, err := stmt.Exec(recipeID, ing.IngredientID, ing.Quantity); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req.ID = recipeID
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(req)
}

func (h *RecipeHandler) GetRecipeIngredients(w http.ResponseWriter, r *http.Request) {
	recipeID := r.URL.Query().Get("recipe_id")
	if recipeID == "" {
		http.Error(w, "recipe_id required", http.StatusBadRequest)
		return
	}

	rows, err := h.DB.Query(`
		SELECT ri.ingredient_id, i.name, ri.quantity, COALESCE(i.unit, '') as unit, COALESCE(i.price, 0) as price, i.is_tracked
		FROM recipe_ingredients ri
		JOIN ingredients i ON ri.ingredient_id = i.id
		WHERE ri.recipe_id = $1
	`, recipeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type IngredientDetail struct {
		IngredientID int     `json:"ingredient_id"`
		Name         string  `json:"name"`
		Quantity     float64 `json:"quantity"`
		Unit         string  `json:"unit"`
		Price        float64 `json:"price"`
		IsTracked    bool    `json:"is_tracked"`
	}

	var ingredients []IngredientDetail
	for rows.Next() {
		var ing IngredientDetail
		if err := rows.Scan(&ing.IngredientID, &ing.Name, &ing.Quantity, &ing.Unit, &ing.Price, &ing.IsTracked); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ingredients = append(ingredients, ing)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ingredients)
}

func (h *RecipeHandler) AddRecipeIngredient(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RecipeID       int     `json:"recipe_id"`
		IngredientID   int     `json:"ingredient_id"`
		IngredientName string  `json:"ingredient_name"`
		Quantity       float64 `json:"quantity"`
		IsTracked      *bool   `json:"is_tracked"` // Optional, default true
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// If no ingredient_id but a name is given, find or create the ingredient
	if req.IngredientID == 0 && req.IngredientName != "" {
		// First try to find existing ingredient by name
		err := h.DB.QueryRow(
			"SELECT id FROM ingredients WHERE LOWER(name) = LOWER($1)",
			req.IngredientName,
		).Scan(&req.IngredientID)

		// If not found, create it
		if err != nil {
			isTracked := true
			if req.IsTracked != nil {
				isTracked = *req.IsTracked
			}

			err = h.DB.QueryRow(
				"INSERT INTO ingredients (name, current_stock, price, is_tracked) VALUES ($1, 0, NULL, $2) RETURNING id",
				req.IngredientName, isTracked,
			).Scan(&req.IngredientID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	if req.IngredientID == 0 {
		http.Error(w, "ingredient_id or ingredient_name required", http.StatusBadRequest)
		return
	}

	_, err := h.DB.Exec(
		`INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity) 
		 VALUES ($1, $2, $3)
		 ON CONFLICT (recipe_id, ingredient_id) 
		 DO UPDATE SET quantity = EXCLUDED.quantity`,
		req.RecipeID, req.IngredientID, req.Quantity,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "added"})
}

func (h *RecipeHandler) RemoveRecipeIngredient(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RecipeID     int `json:"recipe_id"`
		IngredientID int `json:"ingredient_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := h.DB.Exec("DELETE FROM recipe_ingredients WHERE recipe_id = $1 AND ingredient_id = $2", req.RecipeID, req.IngredientID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "removed"})
}

func (h *RecipeHandler) DeleteRecipe(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := h.DB.Exec("DELETE FROM recipes WHERE id = $1", req.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

func (h *RecipeHandler) UpdateRecipeName(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Notes string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	result, err := h.DB.Exec("UPDATE recipes SET name = $1, notes = $2 WHERE id = $3", req.Name, req.Notes, req.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Recipe not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (h *RecipeHandler) UpdateIngredientQuantity(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RecipeID     int     `json:"recipe_id"`
		IngredientID int     `json:"ingredient_id"`
		Quantity     float64 `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Quantity <= 0 {
		http.Error(w, "quantity must be positive", http.StatusBadRequest)
		return
	}

	result, err := h.DB.Exec(
		"UPDATE recipe_ingredients SET quantity = $1 WHERE recipe_id = $2 AND ingredient_id = $3",
		req.Quantity, req.RecipeID, req.IngredientID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Ingredient not found in recipe", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}
