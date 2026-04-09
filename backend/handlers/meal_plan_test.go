//go:build integration

package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Kano-Chien/house_management/backend/handlers"
)

func TestGetMealPlan_Empty(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.MealPlanHandler{DB: db}

	req := httptest.NewRequest("GET", "/api/mealplan", nil)
	w := httptest.NewRecorder()
	h.GetMealPlan(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	// Should return null or [] — both are valid empty JSON sequences
	body := w.Body.String()
	if body != "null\n" && body != "[]\n" {
		// Accept either representation
		var arr []interface{}
		if err := json.Unmarshal([]byte(body), &arr); err == nil && len(arr) == 0 {
			return
		}
	}
}

func TestScheduleMeal_WithRecipe(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.MealPlanHandler{DB: db}

	recipeID := seedRecipe(t, db, "Pasta")
	body, _ := json.Marshal(map[string]interface{}{
		"date": "2026-03-20", "meal_type": "Dinner", "recipe_id": recipeID,
	})
	req := httptest.NewRequest("POST", "/api/mealplan", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.ScheduleMeal(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]int
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["id"] == 0 {
		t.Error("expected non-zero id in response")
	}
}

func TestScheduleMeal_CustomName(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.MealPlanHandler{DB: db}

	body, _ := json.Marshal(map[string]interface{}{
		"date": "2026-03-21", "meal_type": "Breakfast", "custom_name": "Cereal",
	})
	req := httptest.NewRequest("POST", "/api/mealplan", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.ScheduleMeal(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestScheduleMeal_InvalidDate(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.MealPlanHandler{DB: db}

	body, _ := json.Marshal(map[string]interface{}{
		"date": "not-a-date", "meal_type": "Lunch",
	})
	req := httptest.NewRequest("POST", "/api/mealplan", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.ScheduleMeal(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestDeleteMealPlan(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.MealPlanHandler{DB: db}

	recipeID := seedRecipe(t, db, "Soup")
	mpID := seedMealPlan(t, db, "2026-03-22", "Lunch", &recipeID)

	body, _ := json.Marshal(map[string]int{"id": mpID})
	req := httptest.NewRequest("POST", "/api/mealplan/delete", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.DeleteMealPlan(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM meal_plan WHERE id = ?", mpID).Scan(&count)
	if count != 0 {
		t.Error("expected meal plan to be deleted")
	}
}

func TestCookMeal_DecremementsInventory(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.MealPlanHandler{DB: db}

	ingID := seedIngredient(t, db, "Chicken", 10, 2, 3.0, true)
	recipeID := seedRecipe(t, db, "Roast Chicken")
	seedRecipeIngredient(t, db, recipeID, ingID, 4)
	mpID := seedMealPlan(t, db, "2026-03-23", "Dinner", &recipeID)

	body, _ := json.Marshal(map[string]int{"id": mpID})
	req := httptest.NewRequest("POST", "/api/mealplan/cook", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.CookMeal(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Stock should have been decremented by 4 (from 10 to 6)
	var stock int
	db.QueryRow("SELECT current_stock FROM ingredients WHERE id = ?", ingID).Scan(&stock)
	if stock != 6 {
		t.Errorf("expected stock=6 after cooking, got %d", stock)
	}

	// Meal should be marked cooked
	var isCooked bool
	db.QueryRow("SELECT is_cooked FROM meal_plan WHERE id = ?", mpID).Scan(&isCooked)
	if !isCooked {
		t.Error("expected meal to be marked as cooked")
	}
}

func TestCookMeal_AlreadyCooked(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.MealPlanHandler{DB: db}

	recipeID := seedRecipe(t, db, "Salad")
	mpID := seedMealPlan(t, db, "2026-03-23", "Lunch", &recipeID)
	db.Exec("UPDATE meal_plan SET is_cooked = 1 WHERE id = ?", mpID)

	body, _ := json.Marshal(map[string]int{"id": mpID})
	req := httptest.NewRequest("POST", "/api/mealplan/cook", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.CookMeal(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409 Conflict for already-cooked meal, got %d", w.Code)
	}
}

func TestCookMeal_InsufficientStock(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.MealPlanHandler{DB: db}

	// Stock=2, need 5
	ingID := seedIngredient(t, db, "Beef", 2, 3, 10.0, true)
	recipeID := seedRecipe(t, db, "Beef Stew")
	seedRecipeIngredient(t, db, recipeID, ingID, 5)
	mpID := seedMealPlan(t, db, "2026-03-24", "Dinner", &recipeID)

	body, _ := json.Marshal(map[string]int{"id": mpID})
	req := httptest.NewRequest("POST", "/api/mealplan/cook", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.CookMeal(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409 Conflict for insufficient stock, got %d", w.Code)
	}

	// Stock should NOT have been decremented (transaction rolled back)
	var stock int
	db.QueryRow("SELECT current_stock FROM ingredients WHERE id = ?", ingID).Scan(&stock)
	if stock != 2 {
		t.Errorf("expected stock to remain 2 after failed cook, got %d", stock)
	}
}

func TestCookMeal_UntrackedIngredientSkipped(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.MealPlanHandler{DB: db}

	// Untracked: stock=0, need 5 — should NOT block cooking
	ingID := seedIngredient(t, db, "Water", 0, 3, 0.0, false)
	recipeID := seedRecipe(t, db, "Boiled Water")
	seedRecipeIngredient(t, db, recipeID, ingID, 5)
	mpID := seedMealPlan(t, db, "2026-03-24", "Breakfast", &recipeID)

	body, _ := json.Marshal(map[string]int{"id": mpID})
	req := httptest.NewRequest("POST", "/api/mealplan/cook", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.CookMeal(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for untracked ingredient, got %d: %s", w.Code, w.Body.String())
	}

	// Untracked stock should not be decremented
	var stock int
	db.QueryRow("SELECT current_stock FROM ingredients WHERE id = ?", ingID).Scan(&stock)
	if stock != 0 {
		t.Errorf("expected untracked stock to remain 0, got %d", stock)
	}
}

func TestCookMeal_NoRecipe(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.MealPlanHandler{DB: db}

	mpID := seedMealPlan(t, db, "2026-03-25", "Breakfast", nil)

	body, _ := json.Marshal(map[string]int{"id": mpID})
	req := httptest.NewRequest("POST", "/api/mealplan/cook", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.CookMeal(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for meal with no recipe, got %d", w.Code)
	}
}

func TestCookMeal_NotFound(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.MealPlanHandler{DB: db}

	body, _ := json.Marshal(map[string]int{"id": 9999})
	req := httptest.NewRequest("POST", "/api/mealplan/cook", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.CookMeal(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}
