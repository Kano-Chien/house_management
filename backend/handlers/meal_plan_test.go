package handlers

import (
	"database/sql"
	"os"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	schema, err := os.ReadFile("../database/schema_sqlite.sql")
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(string(schema)); err != nil {
		t.Fatalf("apply schema: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func insertTestIngredient(t *testing.T, db *sql.DB, name string, stock int) int {
	t.Helper()
	res, err := db.Exec(
		"INSERT INTO ingredients (name, current_stock, min_stock, is_tracked) VALUES (?, ?, 0, TRUE)",
		name, stock,
	)
	if err != nil {
		t.Fatalf("insert ingredient: %v", err)
	}
	id, _ := res.LastInsertId()
	return int(id)
}

func insertTestBatch(t *testing.T, db *sql.DB, ingredientID, qty int) {
	t.Helper()
	if _, err := db.Exec(
		"INSERT INTO ingredient_batches (ingredient_id, quantity) VALUES (?, ?)",
		ingredientID, qty,
	); err != nil {
		t.Fatalf("insert batch: %v", err)
	}
}

func insertTestRecipe(t *testing.T, db *sql.DB, name string) int {
	t.Helper()
	res, err := db.Exec("INSERT INTO recipes (name) VALUES (?)", name)
	if err != nil {
		t.Fatalf("insert recipe: %v", err)
	}
	id, _ := res.LastInsertId()
	return int(id)
}

func insertTestMeal(t *testing.T, db *sql.DB, date, mealType string, recipeID *int) int {
	t.Helper()
	res, err := db.Exec(
		"INSERT INTO meal_plan (date, meal_type, recipe_id) VALUES (?, ?, ?)",
		date, mealType, recipeID,
	)
	if err != nil {
		t.Fatalf("insert meal: %v", err)
	}
	id, _ := res.LastInsertId()
	return int(id)
}

func insertMealIngredient(t *testing.T, db *sql.DB, mealID, ingredientID, qty int) {
	t.Helper()
	if _, err := db.Exec(
		"INSERT INTO meal_plan_ingredients (meal_plan_id, ingredient_id, quantity) VALUES (?, ?, ?)",
		mealID, ingredientID, qty,
	); err != nil {
		t.Fatalf("insert meal ingredient: %v", err)
	}
}

func TestCookOneMeal_Success(t *testing.T) {
	db := setupTestDB(t)
	h := &MealPlanHandler{DB: db}

	ingID := insertTestIngredient(t, db, "Rice", 10)
	insertTestBatch(t, db, ingID, 10)
	rid := insertTestRecipe(t, db, "Chicken Rice")
	mealID := insertTestMeal(t, db, "2026-05-05", "Lunch", &rid)
	insertMealIngredient(t, db, mealID, ingID, 3)

	missing, err := h.cookOneMeal(mealID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(missing) != 0 {
		t.Errorf("expected no missing ingredients, got %v", missing)
	}

	var isCooked bool
	db.QueryRow("SELECT is_cooked FROM meal_plan WHERE id = ?", mealID).Scan(&isCooked)
	if !isCooked {
		t.Error("meal should be marked cooked")
	}

	var stock int
	db.QueryRow("SELECT current_stock FROM ingredients WHERE id = ?", ingID).Scan(&stock)
	if stock != 7 {
		t.Errorf("expected stock 7, got %d", stock)
	}
}

func TestCookOneMeal_InsufficientStock(t *testing.T) {
	db := setupTestDB(t)
	h := &MealPlanHandler{DB: db}

	ingID := insertTestIngredient(t, db, "Beef", 1)
	insertTestBatch(t, db, ingID, 1)
	rid := insertTestRecipe(t, db, "Beef Stew")
	mealID := insertTestMeal(t, db, "2026-05-05", "Dinner", &rid)
	insertMealIngredient(t, db, mealID, ingID, 5)

	missing, err := h.cookOneMeal(mealID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(missing) == 0 {
		t.Error("expected missing ingredients")
	}

	var isCooked bool
	db.QueryRow("SELECT is_cooked FROM meal_plan WHERE id = ?", mealID).Scan(&isCooked)
	if isCooked {
		t.Error("meal should NOT be marked cooked when stock is insufficient")
	}

	var stock int
	db.QueryRow("SELECT current_stock FROM ingredients WHERE id = ?", ingID).Scan(&stock)
	if stock != 1 {
		t.Errorf("stock should be unchanged at 1, got %d", stock)
	}
}

func TestCookOneMeal_NoIngredients(t *testing.T) {
	db := setupTestDB(t)
	h := &MealPlanHandler{DB: db}

	mealID := insertTestMeal(t, db, "2026-05-05", "Breakfast", nil)

	missing, err := h.cookOneMeal(mealID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(missing) != 0 {
		t.Errorf("expected no missing, got %v", missing)
	}

	var isCooked bool
	db.QueryRow("SELECT is_cooked FROM meal_plan WHERE id = ?", mealID).Scan(&isCooked)
	if !isCooked {
		t.Error("custom meal with no ingredients should still be marked cooked")
	}
}
