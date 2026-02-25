package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

type ShoppingListHandler struct {
	DB *sql.DB
}

type ShoppingItem struct {
	ID            int     `json:"id,omitempty"`
	Name          string  `json:"name"`
	IngredientID  *int    `json:"ingredient_id,omitempty"`
	IsCustom      bool    `json:"is_custom"`
	IsChecked     bool    `json:"is_checked"`
	CurrentStock  int     `json:"current_stock,omitempty"`
	EstimatedCost float64 `json:"estimated_cost,omitempty"`
}

func (h *ShoppingListHandler) GetShoppingList(w http.ResponseWriter, r *http.Request) {
	// Sync: Add auto items for ingredients below min_stock threshold
	_, err := h.DB.Exec(`
		INSERT INTO shopping_list_items (name, ingredient_id, is_custom, is_checked)
		SELECT i.name, i.id, 0, 1
		FROM ingredients i
		WHERE i.current_stock < COALESCE(i.min_stock, 3) AND i.is_tracked = 1
		AND NOT EXISTS (
			SELECT 1 FROM shopping_list_items s 
			WHERE s.ingredient_id = i.id AND s.is_custom = 0
		)
	`)
	if err != nil {
		log.Printf("Error syncing shopping list (insert): %v", err)
		http.Error(w, "Failed to sync shopping list", http.StatusInternalServerError)
		return
	}

	// Remove auto-items that now have sufficient stock or are no longer tracked
	_, err = h.DB.Exec(`
		DELETE FROM shopping_list_items
		WHERE is_custom = 0 AND ingredient_id IN (
			SELECT id FROM ingredients WHERE current_stock >= COALESCE(min_stock, 3) OR is_tracked = 0
		)
	`)
	if err != nil {
		log.Printf("Error syncing shopping list (delete): %v", err)
		http.Error(w, "Failed to sync shopping list", http.StatusInternalServerError)
		return
	}

	// Fetch the actual list, using ingredient_id FK for auto items
	query := `
		SELECT
			s.id,
			s.name,
			s.ingredient_id,
			s.is_custom,
			s.is_checked,
			COALESCE(i.current_stock, 0) as current_stock,
			COALESCE(i.price, 0) as estimated_cost
		FROM shopping_list_items s
		LEFT JOIN ingredients i ON s.ingredient_id = i.id
		ORDER BY s.is_custom ASC, s.id ASC
	`

	rows, err := h.DB.Query(query)
	if err != nil {
		log.Printf("Error querying shopping list: %v", err)
		http.Error(w, "Failed to fetch shopping list", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []ShoppingItem
	for rows.Next() {
		var item ShoppingItem
		if err := rows.Scan(&item.ID, &item.Name, &item.IngredientID, &item.IsCustom, &item.IsChecked, &item.CurrentStock, &item.EstimatedCost); err != nil {
			log.Printf("Error scanning shopping list item: %v", err)
			http.Error(w, "Failed to read shopping list", http.StatusInternalServerError)
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
