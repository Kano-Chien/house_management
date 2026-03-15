//go:build integration

package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Kano-Chien/house_management/backend/handlers"
	"github.com/Kano-Chien/house_management/backend/models"
)

func TestGetRecipes_Empty(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.RecipeHandler{DB: db}

	req := httptest.NewRequest("GET", "/api/recipes", nil)
	w := httptest.NewRecorder()
	h.GetRecipes(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestCreateRecipe_NoIngredients(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.RecipeHandler{DB: db}

	body, _ := json.Marshal(models.Recipe{Name: "Simple Salad"})
	req := httptest.NewRequest("POST", "/api/recipes", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.CreateRecipe(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var r models.Recipe
	json.NewDecoder(w.Body).Decode(&r)
	if r.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if r.Name != "Simple Salad" {
		t.Errorf("expected name=Simple Salad, got %q", r.Name)
	}
}

func TestCreateRecipe_WithIngredients(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.RecipeHandler{DB: db}

	ingID := seedIngredient(t, db, "Tomato", 10, 3, 1.0, true)

	recipe := models.Recipe{
		Name: "Tomato Soup",
		Ingredients: []models.RecipeIngredient{
			{IngredientID: ingID, Quantity: 3},
		},
	}
	body, _ := json.Marshal(recipe)
	req := httptest.NewRequest("POST", "/api/recipes", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.CreateRecipe(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var r models.Recipe
	json.NewDecoder(w.Body).Decode(&r)

	// Verify DB
	var count int
	db.QueryRow("SELECT COUNT(*) FROM recipe_ingredients WHERE recipe_id = ?", r.ID).Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 recipe_ingredient row, got %d", count)
	}
}

func TestGetRecipeIngredients(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.RecipeHandler{DB: db}

	ingID := seedIngredient(t, db, "Garlic", 5, 2, 0.5, true)
	recipeID := seedRecipe(t, db, "Garlic Bread")
	seedRecipeIngredient(t, db, recipeID, ingID, 2)

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/recipes/ingredients?recipe_id=%d", recipeID), nil)
	w := httptest.NewRecorder()
	h.GetRecipeIngredients(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	type IngDetail struct {
		IngredientID int    `json:"ingredient_id"`
		Name         string `json:"name"`
		Quantity     int    `json:"quantity"`
	}
	var ings []IngDetail
	json.NewDecoder(w.Body).Decode(&ings)
	if len(ings) != 1 {
		t.Fatalf("expected 1 ingredient, got %d", len(ings))
	}
	if ings[0].Name != "Garlic" {
		t.Errorf("expected name=Garlic, got %q", ings[0].Name)
	}
}

func TestGetRecipeIngredients_MissingID(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.RecipeHandler{DB: db}

	req := httptest.NewRequest("GET", "/api/recipes/ingredients", nil)
	w := httptest.NewRecorder()
	h.GetRecipeIngredients(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAddRecipeIngredient_ByID(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.RecipeHandler{DB: db}

	ingID := seedIngredient(t, db, "Onion", 8, 3, 0.3, true)
	recipeID := seedRecipe(t, db, "Onion Soup")

	body, _ := json.Marshal(map[string]int{
		"recipe_id": recipeID, "ingredient_id": ingID, "quantity": 2,
	})
	req := httptest.NewRequest("POST", "/api/recipes/ingredients", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.AddRecipeIngredient(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var qty int
	db.QueryRow("SELECT quantity FROM recipe_ingredients WHERE recipe_id = ? AND ingredient_id = ?", recipeID, ingID).Scan(&qty)
	if qty != 2 {
		t.Errorf("expected quantity=2, got %d", qty)
	}
}

func TestAddRecipeIngredient_ByName_NewIngredient(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.RecipeHandler{DB: db}

	recipeID := seedRecipe(t, db, "Mystery Soup")

	body, _ := json.Marshal(map[string]interface{}{
		"recipe_id": recipeID, "ingredient_name": "Dragon Fruit", "quantity": 1,
	})
	req := httptest.NewRequest("POST", "/api/recipes/ingredients", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.AddRecipeIngredient(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Ingredient should have been auto-created
	var count int
	db.QueryRow("SELECT COUNT(*) FROM ingredients WHERE LOWER(name) = 'dragon fruit'").Scan(&count)
	if count != 1 {
		t.Error("expected ingredient Dragon Fruit to be auto-created")
	}
}

func TestAddRecipeIngredient_ByName_ExistingIngredient(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.RecipeHandler{DB: db}

	ingID := seedIngredient(t, db, "Pepper", 5, 3, 0.2, true)
	recipeID := seedRecipe(t, db, "Spicy Noodles")

	body, _ := json.Marshal(map[string]interface{}{
		"recipe_id": recipeID, "ingredient_name": "pepper", "quantity": 1,
	})
	req := httptest.NewRequest("POST", "/api/recipes/ingredients", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.AddRecipeIngredient(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Should have linked to existing ingredient (case-insensitive), not created a new one
	var totalCount int
	db.QueryRow("SELECT COUNT(*) FROM ingredients WHERE LOWER(name) = 'pepper'").Scan(&totalCount)
	if totalCount != 1 {
		t.Errorf("expected 1 pepper ingredient (no duplicate), got %d", totalCount)
	}

	var qty int
	db.QueryRow("SELECT quantity FROM recipe_ingredients WHERE recipe_id = ? AND ingredient_id = ?", recipeID, ingID).Scan(&qty)
	if qty != 1 {
		t.Errorf("expected quantity=1, got %d", qty)
	}
}

func TestRemoveRecipeIngredient(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.RecipeHandler{DB: db}

	ingID := seedIngredient(t, db, "Flour", 10, 3, 1.0, true)
	recipeID := seedRecipe(t, db, "Bread")
	seedRecipeIngredient(t, db, recipeID, ingID, 3)

	body, _ := json.Marshal(map[string]int{"recipe_id": recipeID, "ingredient_id": ingID})
	req := httptest.NewRequest("POST", "/api/recipes/ingredients/remove", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.RemoveRecipeIngredient(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM recipe_ingredients WHERE recipe_id = ? AND ingredient_id = ?", recipeID, ingID).Scan(&count)
	if count != 0 {
		t.Error("expected recipe ingredient to be removed")
	}
}

func TestDeleteRecipe(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.RecipeHandler{DB: db}

	recipeID := seedRecipe(t, db, "Old Recipe")

	body, _ := json.Marshal(map[string]int{"id": recipeID})
	req := httptest.NewRequest("POST", "/api/recipes/delete", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.DeleteRecipe(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM recipes WHERE id = ?", recipeID).Scan(&count)
	if count != 0 {
		t.Error("expected recipe to be deleted")
	}
}

func TestUpdateRecipeName(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.RecipeHandler{DB: db}

	recipeID := seedRecipe(t, db, "Old Name")

	body, _ := json.Marshal(map[string]interface{}{
		"id": recipeID, "name": "New Name", "notes": "Updated notes",
	})
	req := httptest.NewRequest("PUT", "/api/recipes/edit", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.UpdateRecipeName(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var name, notes string
	db.QueryRow("SELECT name, notes FROM recipes WHERE id = ?", recipeID).Scan(&name, &notes)
	if name != "New Name" {
		t.Errorf("expected name=New Name, got %q", name)
	}
	if notes != "Updated notes" {
		t.Errorf("expected notes=Updated notes, got %q", notes)
	}
}

func TestUpdateRecipeName_EmptyName(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.RecipeHandler{DB: db}

	recipeID := seedRecipe(t, db, "My Recipe")

	body, _ := json.Marshal(map[string]interface{}{"id": recipeID, "name": ""})
	req := httptest.NewRequest("PUT", "/api/recipes/edit", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.UpdateRecipeName(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty name, got %d", w.Code)
	}
}

func TestUpdateIngredientQuantity(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.RecipeHandler{DB: db}

	ingID := seedIngredient(t, db, "Oil", 5, 3, 2.0, true)
	recipeID := seedRecipe(t, db, "Stir Fry")
	seedRecipeIngredient(t, db, recipeID, ingID, 1)

	body, _ := json.Marshal(map[string]int{
		"recipe_id": recipeID, "ingredient_id": ingID, "quantity": 4,
	})
	req := httptest.NewRequest("PUT", "/api/recipes/ingredients/edit", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.UpdateIngredientQuantity(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var qty int
	db.QueryRow("SELECT quantity FROM recipe_ingredients WHERE recipe_id = ? AND ingredient_id = ?", recipeID, ingID).Scan(&qty)
	if qty != 4 {
		t.Errorf("expected quantity=4, got %d", qty)
	}
}

func TestUpdateIngredientQuantity_ZeroInvalid(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.RecipeHandler{DB: db}

	recipeID := seedRecipe(t, db, "Test Recipe")
	ingID := seedIngredient(t, db, "Salt", 5, 3, 0.1, true)
	seedRecipeIngredient(t, db, recipeID, ingID, 2)

	body, _ := json.Marshal(map[string]int{
		"recipe_id": recipeID, "ingredient_id": ingID, "quantity": 0,
	})
	req := httptest.NewRequest("PUT", "/api/recipes/ingredients/edit", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.UpdateIngredientQuantity(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for zero quantity, got %d", w.Code)
	}
}

func TestDeleteRecipe_CascadesRecipeIngredients(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.RecipeHandler{DB: db}

	ingID := seedIngredient(t, db, "Basil", 3, 2, 0.5, true)
	recipeID := seedRecipe(t, db, "Pesto")
	seedRecipeIngredient(t, db, recipeID, ingID, 2)

	body, _ := json.Marshal(map[string]int{"id": recipeID})
	req := httptest.NewRequest("POST", "/api/recipes/delete", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.DeleteRecipe(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM recipe_ingredients WHERE recipe_id = ?", recipeID).Scan(&count)
	if count != 0 {
		t.Error("expected recipe_ingredients to be cascade-deleted with recipe")
	}
}
