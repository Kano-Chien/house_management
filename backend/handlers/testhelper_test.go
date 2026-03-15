//go:build integration

package handlers_test

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

const testSchema = `
CREATE TABLE IF NOT EXISTS ingredients (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    current_stock INTEGER DEFAULT 0,
    min_stock INTEGER DEFAULT 3,
    price REAL DEFAULT 0,
    category TEXT DEFAULT 'food',
    is_tracked BOOLEAN DEFAULT 1
);

CREATE TABLE IF NOT EXISTS recipes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    notes TEXT DEFAULT ''
);

CREATE TABLE IF NOT EXISTS recipe_ingredients (
    recipe_id INTEGER,
    ingredient_id INTEGER,
    quantity INTEGER NOT NULL,
    PRIMARY KEY (recipe_id, ingredient_id),
    FOREIGN KEY (recipe_id) REFERENCES recipes(id) ON DELETE CASCADE,
    FOREIGN KEY (ingredient_id) REFERENCES ingredients(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS meal_plan (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date TEXT NOT NULL,
    meal_type TEXT NOT NULL CHECK (meal_type IN ('Breakfast', 'Lunch', 'Dinner')),
    recipe_id INTEGER,
    custom_name TEXT,
    is_cooked BOOLEAN DEFAULT 0,
    FOREIGN KEY (recipe_id) REFERENCES recipes(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS shopping_list_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    ingredient_id INTEGER,
    is_custom BOOLEAN DEFAULT 0,
    is_checked BOOLEAN DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (ingredient_id) REFERENCES ingredients(id) ON DELETE CASCADE
);
`

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		t.Fatalf("failed to enable foreign keys: %v", err)
	}
	if _, err := db.Exec(testSchema); err != nil {
		t.Fatalf("failed to apply schema: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// seed helpers

func seedIngredient(t *testing.T, db *sql.DB, name string, stock, minStock int, price float64, tracked bool) int {
	t.Helper()
	res, err := db.Exec(
		"INSERT INTO ingredients (name, current_stock, min_stock, price, is_tracked) VALUES (?, ?, ?, ?, ?)",
		name, stock, minStock, price, tracked,
	)
	if err != nil {
		t.Fatalf("seedIngredient: %v", err)
	}
	id, _ := res.LastInsertId()
	return int(id)
}

func seedRecipe(t *testing.T, db *sql.DB, name string) int {
	t.Helper()
	res, err := db.Exec("INSERT INTO recipes (name) VALUES (?)", name)
	if err != nil {
		t.Fatalf("seedRecipe: %v", err)
	}
	id, _ := res.LastInsertId()
	return int(id)
}

func seedRecipeIngredient(t *testing.T, db *sql.DB, recipeID, ingredientID, qty int) {
	t.Helper()
	if _, err := db.Exec(
		"INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity) VALUES (?, ?, ?)",
		recipeID, ingredientID, qty,
	); err != nil {
		t.Fatalf("seedRecipeIngredient: %v", err)
	}
}

func seedMealPlan(t *testing.T, db *sql.DB, date, mealType string, recipeID *int) int {
	t.Helper()
	res, err := db.Exec(
		"INSERT INTO meal_plan (date, meal_type, recipe_id) VALUES (?, ?, ?)",
		date, mealType, recipeID,
	)
	if err != nil {
		t.Fatalf("seedMealPlan: %v", err)
	}
	id, _ := res.LastInsertId()
	return int(id)
}
