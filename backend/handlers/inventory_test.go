//go:build integration

package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Kano-Chien/house_management/backend/handlers"
	"github.com/Kano-Chien/house_management/backend/models"
)

func TestGetInventory_Empty(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.InventoryHandler{DB: db}

	req := httptest.NewRequest("GET", "/api/inventory", nil)
	w := httptest.NewRecorder()
	h.GetInventory(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var items []models.Ingredient
	json.NewDecoder(w.Body).Decode(&items)
	if len(items) != 0 {
		t.Errorf("expected empty inventory, got %d items", len(items))
	}
}

func TestGetInventory_WithPlannedConsumption(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.InventoryHandler{DB: db}

	ingID := seedIngredient(t, db, "Rice", 10, 3, 5.0, true)
	recipeID := seedRecipe(t, db, "Fried Rice")
	seedRecipeIngredient(t, db, recipeID, ingID, 3)
	recipeIDPtr := recipeID
	seedMealPlan(t, db, "2026-03-15", "Lunch", &recipeIDPtr)

	req := httptest.NewRequest("GET", "/api/inventory", nil)
	w := httptest.NewRecorder()
	h.GetInventory(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var items []models.Ingredient
	json.NewDecoder(w.Body).Decode(&items)

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].PlannedConsumption != 3 {
		t.Errorf("expected planned_consumption=3, got %d", items[0].PlannedConsumption)
	}
}

func TestAddIngredient(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.InventoryHandler{DB: db}

	body := `{"name":"Salt","current_stock":5,"price":1.5,"category":"food","is_tracked":true,"min_stock":2}`
	req := httptest.NewRequest("POST", "/api/inventory", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	h.AddIngredient(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var ing models.Ingredient
	json.NewDecoder(w.Body).Decode(&ing)
	if ing.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if ing.Name != "Salt" {
		t.Errorf("expected Name=Salt, got %q", ing.Name)
	}
}

func TestAddIngredient_DefaultsCategory(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.InventoryHandler{DB: db}

	body := `{"name":"Sugar","current_stock":3}`
	req := httptest.NewRequest("POST", "/api/inventory", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	h.AddIngredient(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	var ing models.Ingredient
	json.NewDecoder(w.Body).Decode(&ing)
	if ing.Category != "food" {
		t.Errorf("expected default category=food, got %q", ing.Category)
	}
}

func TestUpdateStock(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.InventoryHandler{DB: db}

	id := seedIngredient(t, db, "Eggs", 6, 3, 0.5, true)

	body, _ := json.Marshal(map[string]int{"id": id, "new_stock": 12})
	req := httptest.NewRequest("PUT", "/api/inventory/stock", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.UpdateStock(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var stock int
	db.QueryRow("SELECT current_stock FROM ingredients WHERE id = ?", id).Scan(&stock)
	if stock != 12 {
		t.Errorf("expected stock=12, got %d", stock)
	}
}

func TestUpdateStock_NotFound(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.InventoryHandler{DB: db}

	body, _ := json.Marshal(map[string]int{"id": 999, "new_stock": 5})
	req := httptest.NewRequest("PUT", "/api/inventory/stock", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.UpdateStock(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestEditIngredient(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.InventoryHandler{DB: db}

	id := seedIngredient(t, db, "Milk", 2, 3, 1.0, true)

	body, _ := json.Marshal(map[string]interface{}{
		"id": id, "name": "Full Cream Milk",
		"current_stock": 5, "price": 2.5,
		"category": "daily", "is_tracked": false,
	})
	req := httptest.NewRequest("PUT", "/api/inventory/edit", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.EditIngredient(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var name string
	var category string
	var tracked bool
	db.QueryRow("SELECT name, category, is_tracked FROM ingredients WHERE id = ?", id).Scan(&name, &category, &tracked)
	if name != "Full Cream Milk" {
		t.Errorf("expected name=Full Cream Milk, got %q", name)
	}
	if category != "daily" {
		t.Errorf("expected category=daily, got %q", category)
	}
	if tracked {
		t.Error("expected is_tracked=false")
	}
}

func TestDeleteIngredient(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.InventoryHandler{DB: db}

	id := seedIngredient(t, db, "Butter", 1, 3, 3.0, true)

	body, _ := json.Marshal(map[string]int{"id": id})
	req := httptest.NewRequest("POST", "/api/inventory/delete", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.DeleteIngredient(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM ingredients WHERE id = ?", id).Scan(&count)
	if count != 0 {
		t.Error("expected ingredient to be deleted")
	}
}

func TestDeleteIngredient_NotFound(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.InventoryHandler{DB: db}

	body, _ := json.Marshal(map[string]int{"id": 999})
	req := httptest.NewRequest("POST", "/api/inventory/delete", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.DeleteIngredient(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestAddIngredient_InvalidJSON(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.InventoryHandler{DB: db}

	req := httptest.NewRequest("POST", "/api/inventory", bytes.NewBufferString("not-json"))
	w := httptest.NewRecorder()
	h.AddIngredient(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
