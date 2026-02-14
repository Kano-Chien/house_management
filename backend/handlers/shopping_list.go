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
	Name          string  `json:"name"`
	CurrentStock  float64 `json:"current_stock"`
	Unit          string  `json:"unit"`
	EstimatedCost float64 `json:"estimated_cost"`
}

func (h *ShoppingListHandler) GetShoppingList(w http.ResponseWriter, r *http.Request) {
	// Show tracked items where stock is below threshold (3)
	query := `
		SELECT
			i.name,
			i.current_stock,
			COALESCE(i.unit, '') as unit,
			COALESCE(i.price, 0) as estimated_cost
		FROM ingredients i
		WHERE i.current_stock < 3
		AND i.is_tracked = TRUE
		ORDER BY i.current_stock ASC
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
		if err := rows.Scan(&item.Name, &item.CurrentStock, &item.Unit, &item.EstimatedCost); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		list = append(list, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}
