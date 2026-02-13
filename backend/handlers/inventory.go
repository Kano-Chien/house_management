package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/Kano-Chien/house_management/backend/models"
)

type InventoryHandler struct {
	DB *sql.DB
}

func (h *InventoryHandler) GetInventory(w http.ResponseWriter, r *http.Request) {
	// Query includes calculation for planned consumption based on future meals
	query := `
		SELECT 
			i.id, i.name, i.current_stock, COALESCE(i.unit, '') as unit, i.expiry_date, COALESCE(i.price, 0) as price,
			COALESCE(i.category, 'food') as category,
			COALESCE((
				SELECT SUM(ri.quantity)
				FROM recipe_ingredients ri
				INNER JOIN meal_plan mp ON ri.recipe_id = mp.recipe_id
				WHERE ri.ingredient_id = i.id
			), 0) as planned_consumption
		FROM ingredients i
		GROUP BY i.id
	`
	rows, err := h.DB.Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var inventory []models.Ingredient
	for rows.Next() {
		var i models.Ingredient
		// Use sql.NullFloat64 or similar if needed, but COALESCE handles nulls
		if err := rows.Scan(&i.ID, &i.Name, &i.CurrentStock, &i.Unit, &i.ExpiryDate, &i.Price, &i.Category, &i.PlannedConsumption); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
	err := h.DB.QueryRow(
		"INSERT INTO ingredients (name, current_stock, unit, expiry_date, price, category) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
		i.Name, i.CurrentStock, i.Unit, i.ExpiryDate, i.Price, i.Category,
	).Scan(&i.ID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(i)
}

func (h *InventoryHandler) UpdateStock(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID       int     `json:"id"`
		NewStock float64 `json:"new_stock"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.DB.Exec("UPDATE ingredients SET current_stock = $1 WHERE id = $2", req.NewStock, req.ID)
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
		ID       int     `json:"id"`
		Name     string  `json:"name"`
		Stock    float64 `json:"current_stock"`
		Price    float64 `json:"price"`
		Category string  `json:"category"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Category == "" {
		req.Category = "food"
	}

	result, err := h.DB.Exec(
		"UPDATE ingredients SET name = $1, current_stock = $2, price = $3, category = $4 WHERE id = $5",
		req.Name, req.Stock, req.Price, req.Category, req.ID,
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

	result, err := h.DB.Exec("DELETE FROM ingredients WHERE id = $1", req.ID)
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
