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
			), 0) as planned_consumption,
			(SELECT MIN(b.expiry_date) FROM ingredient_batches b
			 WHERE b.ingredient_id = i.id AND b.expiry_date IS NOT NULL AND b.expiry_date != '') as earliest_expiry
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
		var earliestExpiry sql.NullString
		if err := rows.Scan(&i.ID, &i.Name, &i.CurrentStock, &i.MinStock, &i.Price, &i.Category, &i.IsTracked, &i.PlannedConsumption, &earliestExpiry); err != nil {
			log.Printf("Error scanning inventory item: %v", err)
			http.Error(w, "Failed to read inventory", http.StatusInternalServerError)
			return
		}
		if earliestExpiry.Valid {
			i.EarliestExpiry = &earliestExpiry.String
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

// StockIn adds stock for multiple ingredients at once, creating batch entries with optional expiry dates.
func (h *InventoryHandler) StockIn(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Items []struct {
			IngredientID int    `json:"ingredient_id"`
			Quantity     int    `json:"quantity"`
			ExpiryDate   string `json:"expiry_date"` // YYYY-MM-DD, may be empty
		} `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(req.Items) == 0 {
		http.Error(w, "no items provided", http.StatusBadRequest)
		return
	}

	tx, err := h.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	for _, item := range req.Items {
		if item.Quantity <= 0 {
			continue
		}
		// Update stock
		if _, err := tx.Exec("UPDATE ingredients SET current_stock = current_stock + ? WHERE id = ?", item.Quantity, item.IngredientID); err != nil {
			log.Printf("StockIn update stock error: %v", err)
			http.Error(w, "failed to update stock", http.StatusInternalServerError)
			return
		}
		// Insert batch entry
		var expiryVal any
		if item.ExpiryDate != "" {
			expiryVal = item.ExpiryDate
		}
		if _, err := tx.Exec("INSERT INTO ingredient_batches (ingredient_id, quantity, expiry_date) VALUES (?, ?, ?)", item.IngredientID, item.Quantity, expiryVal); err != nil {
			log.Printf("StockIn insert batch error: %v", err)
			http.Error(w, "failed to insert batch", http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
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
		ID        int    `json:"id"`
		Name      string `json:"name"`
		Price     int    `json:"price"`
		Category  string `json:"category"`
		IsTracked bool   `json:"is_tracked"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Category == "" {
		req.Category = "food"
	}

	result, err := h.DB.Exec(
		"UPDATE ingredients SET name = ?, price = ?, category = ?, is_tracked = ? WHERE id = ?",
		req.Name, req.Price, req.Category, req.IsTracked, req.ID,
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

// GetBatches returns all batches for a given ingredient_id query param.
func (h *InventoryHandler) GetBatches(w http.ResponseWriter, r *http.Request) {
	ingredientID := r.URL.Query().Get("ingredient_id")
	if ingredientID == "" {
		http.Error(w, "ingredient_id required", http.StatusBadRequest)
		return
	}
	rows, err := h.DB.Query(
		"SELECT id, ingredient_id, quantity, COALESCE(expiry_date, '') FROM ingredient_batches WHERE ingredient_id = ? ORDER BY expiry_date ASC, id ASC",
		ingredientID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var batches []models.IngredientBatch
	for rows.Next() {
		var b models.IngredientBatch
		if err := rows.Scan(&b.ID, &b.IngredientID, &b.Quantity, &b.ExpiryDate); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		batches = append(batches, b)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(batches)
}

// UpdateBatch updates a batch's quantity and/or expiry date.
func (h *InventoryHandler) UpdateBatch(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID         int    `json:"id"`
		Quantity   int    `json:"quantity"`
		ExpiryDate string `json:"expiry_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var expiryVal any
	if req.ExpiryDate != "" {
		expiryVal = req.ExpiryDate
	}

	var ingredientID int
	if err := h.DB.QueryRow("SELECT ingredient_id FROM ingredient_batches WHERE id = ?", req.ID).Scan(&ingredientID); err != nil {
		http.Error(w, "batch not found", http.StatusNotFound)
		return
	}

	tx, err := h.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	if _, err := tx.Exec("UPDATE ingredient_batches SET quantity = ?, expiry_date = ? WHERE id = ?", req.Quantity, expiryVal, req.ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Recompute stock as SUM of all batches — no drift possible
	if _, err := tx.Exec(
		"UPDATE ingredients SET current_stock = (SELECT COALESCE(SUM(quantity), 0) FROM ingredient_batches WHERE ingredient_id = ?) WHERE id = ?",
		ingredientID, ingredientID,
	); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

// DeleteBatch removes a batch and reduces ingredient stock accordingly.
func (h *InventoryHandler) DeleteBatch(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var ingredientID int
	if err := h.DB.QueryRow("SELECT ingredient_id FROM ingredient_batches WHERE id = ?", req.ID).Scan(&ingredientID); err != nil {
		http.Error(w, "batch not found", http.StatusNotFound)
		return
	}

	tx, err := h.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM ingredient_batches WHERE id = ?", req.ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Recompute stock as SUM of remaining batches
	if _, err := tx.Exec(
		"UPDATE ingredients SET current_stock = (SELECT COALESCE(SUM(quantity), 0) FROM ingredient_batches WHERE ingredient_id = ?) WHERE id = ?",
		ingredientID, ingredientID,
	); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
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
