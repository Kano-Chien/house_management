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

func TestGetShoppingList_AutoAddsLowStock(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.ShoppingListHandler{DB: db}

	// stock=1, min_stock=3 → below threshold, tracked
	seedIngredient(t, db, "Eggs", 1, 3, 0.5, true)

	req := httptest.NewRequest("GET", "/api/shopping-list", nil)
	w := httptest.NewRecorder()
	h.GetShoppingList(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var items []handlers.ShoppingItem
	json.NewDecoder(w.Body).Decode(&items)

	if len(items) != 1 {
		t.Fatalf("expected 1 auto-item, got %d", len(items))
	}
	if items[0].Name != "Eggs" {
		t.Errorf("expected Name=Eggs, got %q", items[0].Name)
	}
	if items[0].IsCustom {
		t.Error("expected auto item to have is_custom=false")
	}
}

func TestGetShoppingList_DoesNotAddSufficientStock(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.ShoppingListHandler{DB: db}

	// stock=5, min_stock=3 → above threshold
	seedIngredient(t, db, "Milk", 5, 3, 1.0, true)

	req := httptest.NewRequest("GET", "/api/shopping-list", nil)
	w := httptest.NewRecorder()
	h.GetShoppingList(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var items []handlers.ShoppingItem
	json.NewDecoder(w.Body).Decode(&items)
	if len(items) != 0 {
		t.Errorf("expected no items for sufficient stock, got %d", len(items))
	}
}

func TestGetShoppingList_DoesNotAddUntrackedIngredient(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.ShoppingListHandler{DB: db}

	// stock=0 but is_tracked=false → should not appear
	seedIngredient(t, db, "Water", 0, 3, 0.0, false)

	req := httptest.NewRequest("GET", "/api/shopping-list", nil)
	w := httptest.NewRecorder()
	h.GetShoppingList(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var items []handlers.ShoppingItem
	json.NewDecoder(w.Body).Decode(&items)
	if len(items) != 0 {
		t.Errorf("expected no items for untracked ingredient, got %d", len(items))
	}
}

func TestGetShoppingList_RemovesItemWhenStockRestored(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.ShoppingListHandler{DB: db}

	// Create ingredient with low stock and an existing auto shopping item
	ingID := seedIngredient(t, db, "Butter", 1, 3, 2.0, true)
	db.Exec("INSERT INTO shopping_list_items (name, ingredient_id, is_custom, is_checked) VALUES (?, ?, 0, 1)", "Butter", ingID)

	// Now restore stock above threshold
	db.Exec("UPDATE ingredients SET current_stock = 5 WHERE id = ?", ingID)

	req := httptest.NewRequest("GET", "/api/shopping-list", nil)
	w := httptest.NewRecorder()
	h.GetShoppingList(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var items []handlers.ShoppingItem
	json.NewDecoder(w.Body).Decode(&items)
	for _, item := range items {
		if item.Name == "Butter" && !item.IsCustom {
			t.Error("expected Butter auto-item to be removed when stock is sufficient")
		}
	}
}

func TestGetShoppingList_NoDuplicateAutoItems(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.ShoppingListHandler{DB: db}

	ingID := seedIngredient(t, db, "Coffee", 0, 3, 5.0, true)
	// Pre-insert an auto item (as if it was already added)
	db.Exec("INSERT INTO shopping_list_items (name, ingredient_id, is_custom, is_checked) VALUES (?, ?, 0, 1)", "Coffee", ingID)

	// Call GetShoppingList twice — should not create duplicate
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/api/shopping-list", nil)
		w := httptest.NewRecorder()
		h.GetShoppingList(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("call %d: expected 200, got %d", i+1, w.Code)
		}
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM shopping_list_items WHERE ingredient_id = ? AND is_custom = 0", ingID).Scan(&count)
	if count != 1 {
		t.Errorf("expected exactly 1 auto-item for Coffee, got %d (duplicate created)", count)
	}
}

func TestAddCustomItem(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.ShoppingListHandler{DB: db}

	body, _ := json.Marshal(map[string]string{"name": "Dish Soap"})
	req := httptest.NewRequest("POST", "/api/shopping-list", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.AddCustomItem(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM shopping_list_items WHERE name = 'Dish Soap' AND is_custom = 1").Scan(&count)
	if count != 1 {
		t.Error("expected custom item Dish Soap to be created")
	}
}

func TestAddCustomItem_EmptyName(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.ShoppingListHandler{DB: db}

	body, _ := json.Marshal(map[string]string{"name": ""})
	req := httptest.NewRequest("POST", "/api/shopping-list", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.AddCustomItem(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty name, got %d", w.Code)
	}
}

func TestToggleItemChecked(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.ShoppingListHandler{DB: db}

	res, _ := db.Exec("INSERT INTO shopping_list_items (name, is_custom, is_checked) VALUES ('Bread', 1, 1)")
	itemID, _ := res.LastInsertId()

	body, _ := json.Marshal(map[string]interface{}{"id": itemID, "is_checked": false})
	req := httptest.NewRequest("PUT", "/api/shopping-list/check", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.ToggleItemChecked(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var checked bool
	db.QueryRow("SELECT is_checked FROM shopping_list_items WHERE id = ?", itemID).Scan(&checked)
	if checked {
		t.Error("expected is_checked=false after toggle")
	}
}

func TestDeleteItem(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.ShoppingListHandler{DB: db}

	res, _ := db.Exec("INSERT INTO shopping_list_items (name, is_custom, is_checked) VALUES ('Juice', 1, 0)")
	itemID, _ := res.LastInsertId()

	body, _ := json.Marshal(map[string]interface{}{"id": itemID})
	req := httptest.NewRequest("POST", "/api/shopping-list/delete", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.DeleteItem(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM shopping_list_items WHERE id = ?", itemID).Scan(&count)
	if count != 0 {
		t.Error("expected item to be deleted")
	}
}

func TestGetShoppingList_CustomItemNotRemovedBySync(t *testing.T) {
	db := newTestDB(t)
	h := &handlers.ShoppingListHandler{DB: db}

	// A custom item with no ingredient_id should never be removed by the auto-sync
	ingID := seedIngredient(t, db, "Napkins", 10, 3, 0.1, true)
	db.Exec("INSERT INTO shopping_list_items (name, ingredient_id, is_custom, is_checked) VALUES (?, ?, 1, 0)", "Napkins (extra)", ingID)

	req := httptest.NewRequest("GET", "/api/shopping-list", nil)
	w := httptest.NewRecorder()
	h.GetShoppingList(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var items []handlers.ShoppingItem
	json.NewDecoder(w.Body).Decode(&items)

	found := false
	for _, item := range items {
		if item.Name == "Napkins (extra)" && item.IsCustom {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected custom item to survive auto-sync even when ingredient stock is sufficient")
	}
}
