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
	ID            int     `json:"id,omitempty"`
	Name          string  `json:"name"`
	IsCustom      bool    `json:"is_custom"`
	IsChecked     bool    `json:"is_checked"`
	CurrentStock  float64 `json:"current_stock,omitempty"`
	Unit          string  `json:"unit,omitempty"`
	EstimatedCost float64 `json:"estimated_cost,omitempty"`
}

func (h *ShoppingListHandler) GetShoppingList(w http.ResponseWriter, r *http.Request) {
	// Let's first sync: Add anything that's current_stock < 3 and is_tracked=true that ISN'T already in shopping_list_items
	// This ensures our DB shopping list is always populated with auto items correctly.
	_, err := h.DB.Exec(`
		INSERT INTO shopping_list_items (name, is_custom, is_checked)
		SELECT i.name, 0, 1
		FROM ingredients i
		WHERE i.current_stock < 3 AND i.is_tracked = 1
		AND NOT EXISTS (
			SELECT 1 FROM shopping_list_items s 
			WHERE LOWER(s.name) = LOWER(i.name) AND s.is_custom = 0
		)
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Conversely, if an auto-item now has stock >= 3 or is_tracked=false, we might want to remove it.
	_, err = h.DB.Exec(`
		DELETE FROM shopping_list_items
		WHERE is_custom = 0 AND name IN (
			SELECT name FROM ingredients WHERE current_stock >= 3 OR is_tracked = 0
		)
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Now fetch the actual list from shopping_list_items
	// We left join ingredients so we can still get cost and stock info if it exists
	query := `
		SELECT
			s.id,
			s.name,
			s.is_custom,
			s.is_checked,
			COALESCE(i.current_stock, 0) as current_stock,
			COALESCE(i.unit, '') as unit,
			COALESCE(i.price, 0) as estimated_cost
		FROM shopping_list_items s
		LEFT JOIN ingredients i ON LOWER(s.name) = LOWER(i.name)
		ORDER BY s.is_custom ASC, s.id ASC
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
		if err := rows.Scan(&item.ID, &item.Name, &item.IsCustom, &item.IsChecked, &item.CurrentStock, &item.Unit, &item.EstimatedCost); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		list = append(list, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *ShoppingListHandler) AddCustomItem(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}

	_, err := h.DB.Exec("INSERT INTO shopping_list_items (name, is_custom, is_checked) VALUES (?, 1, 1)", req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "added"})
}

func (h *ShoppingListHandler) ToggleItemChecked(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID        int  `json:"id"`
		IsChecked bool `json:"is_checked"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := h.DB.Exec("UPDATE shopping_list_items SET is_checked = ? WHERE id = ?", req.IsChecked, req.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (h *ShoppingListHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := h.DB.Exec("DELETE FROM shopping_list_items WHERE id = ?", req.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}
