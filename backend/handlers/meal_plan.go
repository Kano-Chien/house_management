package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Kano-Chien/house_management/backend/models"
)

type MealPlanHandler struct {
	DB *sql.DB
}

func (h *MealPlanHandler) GetMealPlan(w http.ResponseWriter, r *http.Request) {
	// Optional: Filter by date range query params ?start=...&end=...
	rows, err := h.DB.Query(`
		SELECT mp.id, mp.date, mp.meal_type, mp.recipe_id, r.name 
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
		var rName sql.NullString
		if err := rows.Scan(&mp.ID, &mp.Date, &mp.MealType, &mp.RecipeID, &rName); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
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
	// Fix date parsing if JSON sends string, but let's assume standard ISO8601 handled by Go's JSON parser to time.Time if format matches
	// Or use a custom struct for decoding
	type Request struct {
		Date     string `json:"date"` // YYYY-MM-DD
		MealType string `json:"meal_type"`
		RecipeID *int   `json:"recipe_id"`
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

	var id int
	err = h.DB.QueryRow("INSERT INTO meal_plan (date, meal_type, recipe_id) VALUES ($1, $2, $3) RETURNING id", date, input.MealType, input.RecipeID).Scan(&id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int{"id": id})
}

func (h *MealPlanHandler) DeleteMealPlan(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err := h.DB.Exec("DELETE FROM meal_plan WHERE id = $1", req.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}
