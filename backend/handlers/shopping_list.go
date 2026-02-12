package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

type ShoppingListHandler struct {
	DB *sql.DB
}

type ShoppingItem struct {
	Name           string  `json:"name"`
	QuantityNeeded float64 `json:"quantity_needed"`
	Unit           string  `json:"unit"`
	EstimatedCost  float64 `json:"estimated_cost"`
}

func (h *ShoppingListHandler) GetShoppingList(w http.ResponseWriter, r *http.Request) {
	// Query to find items where required quantity (from meal plan) > current stock
	// This uses a CTE or subquery to sum up requirements first
	query := `
		WITH RequiredIngredients AS (
			SELECT 
				ri.ingredient_id,
				SUM(ri.quantity) as total_required
			FROM recipe_ingredients ri
			JOIN meal_plan mp ON ri.recipe_id = mp.recipe_id
			WHERE mp.date >= CURRENT_DATE
			GROUP BY ri.ingredient_id
		)
		SELECT 
			i.name,
			(req.total_required - i.current_stock) as quantity_needed,
			i.unit,
			((req.total_required - i.current_stock) * i.price) as estimated_cost
		FROM ingredients i
		JOIN RequiredIngredients req ON i.id = req.ingredient_id
		WHERE i.current_stock < req.total_required
	`

	rows, err := h.DB.Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []ShoppingItem
	for rows.Next() {
		var item ShoppingItem
		if err := rows.Scan(&item.Name, &item.QuantityNeeded, &item.Unit, &item.EstimatedCost); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		list = append(list, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}
