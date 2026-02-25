package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/Kano-Chien/house_management/backend/models"
)

type InventoryHandler struct {
	DB *sql.DB
}

func (h *InventoryHandler) GetInventory(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT
			i.id, i.name, i.current_stock, COALESCE(i.min_stock, 3) as min_stock,
			COALESCE(i.price, 0) as price,
			COALESCE(i.category, 'food') as category,
			i.is_tracked,
			COALESCE((
				SELECT SUM(ri.quantity)
				FROM recipe_ingredients ri
				INNER JOIN meal_plan mp ON ri.recipe_id = mp.recipe_id
				WHERE ri.ingredient_id = i.id AND (mp.is_cooked = FALSE OR mp.is_cooked = 0)
			), 0) as planned_consumption
		FROM ingredients i
		GROUP BY i.id
	`
	rows, err := h.DB.Query(query)
	if err != nil {
		log.Printf("Error querying inventory: %v", err)
		http.Error(w, "Failed to fetch inventory", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var inventory []models.Ingredient
	for rows.Next() {
		var i models.Ingredient
		if err := rows.Scan(&i.ID, &i.Name, &i.CurrentStock, &i.MinStock, &i.Price, &i.Category, &i.IsTracked, &i.PlannedConsumption); err != nil {
			log.Printf("Error scanning inventory item: %v", err)
			http.Error(w, "Failed to read inventory", http.StatusInternalServerError)
			return
		}
		inventory = append(inventory, i)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(inventory)
}

func (h *InventoryHandler) AddIngredient(w http.ResponseWriter, r *http.Request) {
	var i models.Ingredient
	if err := json.NewDecoder(r.Body).Decode(&i); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if i.Category == "" {
		i.Category = "food"
	}
	res, err := h.DB.Exec(
		"INSERT INTO ingredients (name, current_stock, price, category, is_tracked, min_stock) VALUES (?, ?, ?, ?, ?, ?)",
		i.Name, i.CurrentStock, i.Price, i.Category, i.IsTracked, i.MinStock,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id, err := res.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	i.ID = int(id)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(i)
}

func (h *InventoryHandler) UpdateStock(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID       int `json:"id"`
		NewStock int `json:"new_stock"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.DB.Exec("UPDATE ingredients SET current_stock = ? WHERE id = ?", req.NewStock, req.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Ingredient not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (h *InventoryHandler) EditIngredient(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID        int     `json:"id"`
		Name      string  `json:"name"`
		Stock     int     `json:"current_stock"`
		Price     float64 `json:"price"`
		Category  string  `json:"category"`
		IsTracked bool    `json:"is_tracked"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Category == "" {
		req.Category = "food"
	}

	result, err := h.DB.Exec(
		"UPDATE ingredients SET name = ?, current_stock = ?, price = ?, category = ?, is_tracked = ? WHERE id = ?",
		req.Name, req.Stock, req.Price, req.Category, req.IsTracked, req.ID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Ingredient not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (h *InventoryHandler) DeleteIngredient(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.DB.Exec("DELETE FROM ingredients WHERE id = ?", req.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Ingredient not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}
