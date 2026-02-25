package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
)

func (h *RecipeHandler) ReplaceRecipeIngredient(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RecipeID        int    `json:"recipe_id"`
		OldIngredientID int    `json:"old_ingredient_id"`
		NewName         string `json:"new_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	req.NewName = strings.TrimSpace(req.NewName)
	if req.NewName == "" || req.RecipeID == 0 || req.OldIngredientID == 0 {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	tx, err := h.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// 1. Get existing quantity for the old ingredient in this recipe
	var quantity int
	err = tx.QueryRow("SELECT quantity FROM recipe_ingredients WHERE recipe_id = ? AND ingredient_id = ?", req.RecipeID, req.OldIngredientID).Scan(&quantity)
	if err != nil {
		http.Error(w, "ingredient not found in recipe", http.StatusNotFound)
		return
	}

	// 2. Check if the new name already exists in the inventory
	var existingIngredientID int
	err = tx.QueryRow("SELECT id FROM ingredients WHERE LOWER(name) = LOWER(?)", req.NewName).Scan(&existingIngredientID)

	if err == sql.ErrNoRows {
		// Case A: The new name does NOT exist in the inventory.
		// Rename the ingredient globally — ingredients are a shared concept across all recipes.
		_, err := tx.Exec("UPDATE ingredients SET name = ? WHERE id = ?", req.NewName, req.OldIngredientID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		// Case B: The new name DOES exist in the inventory (and is not exactly the same item, though it might be).
		// If it's the exact same item ID, we just do nothing or just return ok.
		if existingIngredientID != req.OldIngredientID {
			// We need to change the recipe to use the existing ingredient.

			// 1. Delete the old reference in recipe_ingredients
			_, err = tx.Exec("DELETE FROM recipe_ingredients WHERE recipe_id = ? AND ingredient_id = ?", req.RecipeID, req.OldIngredientID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// 2. Insert or update the new reference in recipe_ingredients
			_, err = tx.Exec(`
				INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity) 
				VALUES (?, ?, ?)
				ON CONFLICT (recipe_id, ingredient_id) 
				DO UPDATE SET quantity = recipe_ingredients.quantity + EXCLUDED.quantity
			`, req.RecipeID, existingIngredientID, quantity)
			if err != nil {
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
