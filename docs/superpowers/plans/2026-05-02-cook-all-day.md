# Cook All Day Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a "Cook All Day" button to each meal plan day card that cooks all uncooked meals in one action, decrements stock for each, and shows a result modal with cooked/failed breakdown.

**Architecture:** New `POST /api/mealplan/cook-day` endpoint processes each meal independently (one transaction per meal for partial-success semantics). A shared `cookOneMeal` helper encapsulates the mark-cooked + FEFO stock-decrement logic. Frontend adds confirm dialog and result modal as inline `v-if` blocks in `MealPlanner.vue`.

**Tech Stack:** Go (stdlib net/http, modernc.org/sqlite), Vue 3 Composition API, TailwindCSS

---

## File Map

| File | Change |
|---|---|
| `backend/handlers/meal_plan.go` | Add `cookOneMeal` helper + `CookAllDay` handler |
| `backend/handlers/meal_plan_test.go` | New — Go tests for `cookOneMeal` and `CookAllDay` |
| `backend/main.go` | Register `POST /api/mealplan/cook-day` route (line ~231) |
| `frontend/src/components/MealPlanner.vue` | New state/functions, updated day card footer, confirm dialog, result modal |

---

### Task 1: Go test infrastructure

**Files:**
- Create: `backend/handlers/meal_plan_test.go`

- [ ] **Step 1: Create the test file with setup helper and fixture functions**

Create `backend/handlers/meal_plan_test.go`:

```go
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
```

- [ ] **Step 2: Verify it compiles**

```bash
cd backend && go test ./handlers/ -run TestNothing -v 2>&1 | head -5
```

Expected: no compile errors (output will be `testing: warning: no tests to run` or similar).

- [ ] **Step 3: Commit**

```bash
git add backend/handlers/meal_plan_test.go
git commit -m "test: add test db helpers for meal plan handler"
```

---

### Task 2: `cookOneMeal` helper with tests

**Files:**
- Modify: `backend/handlers/meal_plan.go`
- Modify: `backend/handlers/meal_plan_test.go`

- [ ] **Step 1: Write failing tests**

Append to `backend/handlers/meal_plan_test.go`:

```go
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
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd backend && go test ./handlers/ -run TestCookOneMeal -v 2>&1 | head -10
```

Expected: compile error — `cookOneMeal undefined`.

- [ ] **Step 3: Add `cookOneMeal` to `meal_plan.go`**

Add this method after the closing `}` of `CookMeal` (after line 403):

```go
// cookOneMeal marks one meal as cooked and decrements stock via FEFO batches.
// Returns missing ingredient strings if stock is insufficient (meal is NOT cooked).
// Returns nil, nil on success or when the meal has no tracked ingredients.
func (h *MealPlanHandler) cookOneMeal(mealID int) (missing []string, err error) {
	tx, err := h.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var recipeID sql.NullInt64
	var isCooked bool
	err = tx.QueryRow("SELECT recipe_id, COALESCE(is_cooked, FALSE) FROM meal_plan WHERE id = ?", mealID).Scan(&recipeID, &isCooked)
	if err != nil {
		return nil, err
	}
	if isCooked {
		return nil, nil
	}

	if _, err = tx.Exec("UPDATE meal_plan SET is_cooked = TRUE WHERE id = ?", mealID); err != nil {
		return nil, err
	}

	type ingredientUpdate struct {
		ID   int
		Name string
		Qty  int
	}

	scanRows := func(rows *sql.Rows) ([]ingredientUpdate, []string, error) {
		var upd []ingredientUpdate
		var miss []string
		for rows.Next() {
			var u ingredientUpdate
			var currentStock int
			if err := rows.Scan(&u.ID, &u.Name, &u.Qty, &currentStock); err != nil {
				rows.Close()
				return nil, nil, err
			}
			if currentStock < u.Qty {
				miss = append(miss, fmt.Sprintf("%s (Need: %d, Have: %d)", u.Name, u.Qty, currentStock))
			}
			upd = append(upd, u)
		}
		rows.Close()
		return upd, miss, rows.Err()
	}

	rows, err := tx.Query(`
		SELECT mpi.ingredient_id, i.name, mpi.quantity, i.current_stock
		FROM meal_plan_ingredients mpi
		JOIN ingredients i ON mpi.ingredient_id = i.id
		WHERE mpi.meal_plan_id = ? AND i.is_tracked = TRUE
	`, mealID)
	if err != nil {
		return nil, err
	}
	updates, missIngredients, err := scanRows(rows)
	if err != nil {
		return nil, err
	}

	if len(updates) == 0 && recipeID.Valid {
		if _, err = tx.Exec(`
			INSERT INTO meal_plan_ingredients (meal_plan_id, ingredient_id, quantity)
			SELECT ?, ri.ingredient_id, ri.quantity
			FROM recipe_ingredients ri WHERE ri.recipe_id = ?
		`, mealID, recipeID.Int64); err != nil {
			return nil, err
		}
		rows, err = tx.Query(`
			SELECT mpi.ingredient_id, i.name, mpi.quantity, i.current_stock
			FROM meal_plan_ingredients mpi
			JOIN ingredients i ON mpi.ingredient_id = i.id
			WHERE mpi.meal_plan_id = ? AND i.is_tracked = TRUE
		`, mealID)
		if err != nil {
			return nil, err
		}
		updates, missIngredients, err = scanRows(rows)
		if err != nil {
			return nil, err
		}
	}

	if len(missIngredients) > 0 {
		return missIngredients, nil
	}

	for _, u := range updates {
		remaining := u.Qty
		batchRows, err := tx.Query(`
			SELECT id, quantity FROM ingredient_batches
			WHERE ingredient_id = ?
			ORDER BY CASE WHEN expiry_date IS NULL OR expiry_date = '' THEN 1 ELSE 0 END,
			         expiry_date ASC, id ASC
		`, u.ID)
		if err != nil {
			return nil, err
		}
		type batchItem struct{ ID, Qty int }
		var batches []batchItem
		for batchRows.Next() {
			var b batchItem
			if err := batchRows.Scan(&b.ID, &b.Qty); err != nil {
				batchRows.Close()
				return nil, err
			}
			batches = append(batches, b)
		}
		batchRows.Close()

		for _, b := range batches {
			if remaining <= 0 {
				break
			}
			consume := b.Qty
			if consume > remaining {
				consume = remaining
			}
			remaining -= consume
			newQty := b.Qty - consume
			if newQty == 0 {
				if _, err := tx.Exec("DELETE FROM ingredient_batches WHERE id = ?", b.ID); err != nil {
					return nil, err
				}
			} else {
				if _, err := tx.Exec("UPDATE ingredient_batches SET quantity = ? WHERE id = ?", newQty, b.ID); err != nil {
					return nil, err
				}
			}
		}
		if _, err := tx.Exec(
			"UPDATE ingredients SET current_stock = (SELECT COALESCE(SUM(quantity), 0) FROM ingredient_batches WHERE ingredient_id = ?) WHERE id = ?",
			u.ID, u.ID,
		); err != nil {
			return nil, err
		}
	}

	return nil, tx.Commit()
}
```

- [ ] **Step 4: Run tests**

```bash
cd backend && go test ./handlers/ -run TestCookOneMeal -v
```

Expected:
```
--- PASS: TestCookOneMeal_Success (0.00s)
--- PASS: TestCookOneMeal_InsufficientStock (0.00s)
--- PASS: TestCookOneMeal_NoIngredients (0.00s)
PASS
```

- [ ] **Step 5: Commit**

```bash
git add backend/handlers/meal_plan.go backend/handlers/meal_plan_test.go
git commit -m "feat(backend): add cookOneMeal helper with tests"
```

---

### Task 3: `CookAllDay` handler with tests

**Files:**
- Modify: `backend/handlers/meal_plan.go`
- Modify: `backend/handlers/meal_plan_test.go`

- [ ] **Step 1: Write failing test**

Append to `backend/handlers/meal_plan_test.go` (add `"bytes"`, `"encoding/json"`, `"net/http"`, `"net/http/httptest"` to the import block at the top of the file first):

```go
func TestCookAllDay_PartialSuccess(t *testing.T) {
	db := setupTestDB(t)
	h := &MealPlanHandler{DB: db}

	// Meal A: sufficient stock
	ingA := insertTestIngredient(t, db, "Rice", 10)
	insertTestBatch(t, db, ingA, 10)
	ridA := insertTestRecipe(t, db, "Chicken Rice")
	mealA := insertTestMeal(t, db, "2026-05-05", "Lunch", &ridA)
	insertMealIngredient(t, db, mealA, ingA, 2)

	// Meal B: insufficient stock
	ingB := insertTestIngredient(t, db, "Beef", 0)
	ridB := insertTestRecipe(t, db, "Beef Stew")
	mealB := insertTestMeal(t, db, "2026-05-05", "Dinner", &ridB)
	insertMealIngredient(t, db, mealB, ingB, 3)

	body, _ := json.Marshal(map[string]string{"date": "2026-05-05"})
	req := httptest.NewRequest(http.MethodPost, "/api/mealplan/cook-day", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CookAllDay(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var result struct {
		Cooked []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"cooked"`
		Failed []struct {
			ID      int      `json:"id"`
			Name    string   `json:"name"`
			Missing []string `json:"missing"`
		} `json:"failed"`
	}
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(result.Cooked) != 1 || result.Cooked[0].ID != mealA {
		t.Errorf("expected cooked=[%d], got %+v", mealA, result.Cooked)
	}
	if len(result.Failed) != 1 || result.Failed[0].ID != mealB {
		t.Errorf("expected failed=[%d], got %+v", mealB, result.Failed)
	}
	if len(result.Failed[0].Missing) == 0 {
		t.Error("expected missing ingredient details for failed meal")
	}
}

func TestCookAllDay_SkipsAlreadyCooked(t *testing.T) {
	db := setupTestDB(t)
	h := &MealPlanHandler{DB: db}

	rid := insertTestRecipe(t, db, "Oatmeal")
	mealID := insertTestMeal(t, db, "2026-05-06", "Breakfast", &rid)
	db.Exec("UPDATE meal_plan SET is_cooked = TRUE WHERE id = ?", mealID)

	body, _ := json.Marshal(map[string]string{"date": "2026-05-06"})
	req := httptest.NewRequest(http.MethodPost, "/api/mealplan/cook-day", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CookAllDay(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var result struct {
		Cooked []interface{} `json:"cooked"`
		Failed []interface{} `json:"failed"`
	}
	json.NewDecoder(w.Body).Decode(&result)
	if len(result.Cooked) != 0 || len(result.Failed) != 0 {
		t.Errorf("expected empty results for already-cooked day, got cooked=%d failed=%d", len(result.Cooked), len(result.Failed))
	}
}
```

- [ ] **Step 2: Update the import block in `meal_plan_test.go`**

Change the import block at the top of the file to:

```go
import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	_ "modernc.org/sqlite"
)
```

- [ ] **Step 3: Run tests to verify they fail**

```bash
cd backend && go test ./handlers/ -run TestCookAllDay -v 2>&1 | head -10
```

Expected: compile error — `CookAllDay undefined`.

- [ ] **Step 4: Add `CookAllDay` handler to `meal_plan.go`**

Add after `cookOneMeal`:

```go
func (h *MealPlanHandler) CookAllDay(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Date string `json:"date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if _, err := time.Parse("2006-01-02", req.Date); err != nil {
		http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	rows, err := h.DB.Query(`
		SELECT mp.id, COALESCE(mp.custom_name, r.name, 'Unnamed')
		FROM meal_plan mp
		LEFT JOIN recipes r ON mp.recipe_id = r.id
		WHERE mp.date = ? AND COALESCE(mp.is_cooked, FALSE) = FALSE
	`, req.Date)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	type mealInfo struct {
		ID   int
		Name string
	}
	var meals []mealInfo
	for rows.Next() {
		var m mealInfo
		if err := rows.Scan(&m.ID, &m.Name); err != nil {
			rows.Close()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		meals = append(meals, m)
	}
	rows.Close()

	type cookedResult struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	type failedResult struct {
		ID      int      `json:"id"`
		Name    string   `json:"name"`
		Missing []string `json:"missing"`
	}

	cooked := []cookedResult{}
	failed := []failedResult{}

	for _, m := range meals {
		missIngredients, err := h.cookOneMeal(m.ID)
		if err != nil {
			failed = append(failed, failedResult{ID: m.ID, Name: m.Name, Missing: []string{err.Error()}})
			continue
		}
		if len(missIngredients) > 0 {
			failed = append(failed, failedResult{ID: m.ID, Name: m.Name, Missing: missIngredients})
		} else {
			cooked = append(cooked, cookedResult{ID: m.ID, Name: m.Name})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"cooked": cooked,
		"failed": failed,
	})
}
```

- [ ] **Step 5: Run all handler tests**

```bash
cd backend && go test ./handlers/ -v
```

Expected: all tests PASS.

- [ ] **Step 6: Commit**

```bash
git add backend/handlers/meal_plan.go backend/handlers/meal_plan_test.go
git commit -m "feat(backend): add CookAllDay handler with tests"
```

---

### Task 4: Register route in `main.go`

**Files:**
- Modify: `backend/main.go`

- [ ] **Step 1: Add the route after the existing `/api/mealplan/cook` block (around line 231)**

After the closing `})` of the `/api/mealplan/cook` handler, add:

```go
mux.HandleFunc("/api/mealplan/cook-day", func(w http.ResponseWriter, r *http.Request) {
    if r.Method == "POST" {
        mealPlanHandler.CookAllDay(w, r)
    } else {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
})
```

- [ ] **Step 2: Build to verify no compile errors**

```bash
cd backend && go build ./...
```

Expected: exits with code 0, no output.

- [ ] **Step 3: Smoke test**

```bash
cd backend && go run main.go &
sleep 1
curl -s -X POST http://localhost:8080/api/mealplan/cook-day \
  -H "Content-Type: application/json" \
  -d '{"date":"2099-01-01"}'
kill %1
```

Expected output: `{"cooked":[],"failed":[]}`

- [ ] **Step 4: Commit**

```bash
git add backend/main.go
git commit -m "feat(backend): register POST /api/mealplan/cook-day route"
```

---

### Task 5: Frontend — state and API functions

**Files:**
- Modify: `frontend/src/components/MealPlanner.vue`

- [ ] **Step 1: Add three new refs**

After the line `const editingNewMealKey = ref(null)` (line 321), add:

```js
const confirmCookAllDayDate = ref(null)
const cookAllDayResult = ref(null)
const isCookingAll = ref(false)
```

- [ ] **Step 2: Add the three new functions**

After the `confirmCookMealItem` function (after line 685), add:

```js
const cookAllDay = (dateStr) => {
  confirmCookAllDayDate.value = dateStr
}

const confirmCookAllDayAction = async () => {
  const date = confirmCookAllDayDate.value
  confirmCookAllDayDate.value = null
  isCookingAll.value = true
  try {
    const res = await fetch('/api/mealplan/cook-day', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ date })
    })
    if (!res.ok) {
      toast('Failed to cook meals', 'error')
      return
    }
    const result = await res.json()
    await fetchMealPlan()
    if (result.cooked.length === 0 && result.failed.length === 0) {
      toast('Nothing to cook')
    } else {
      cookAllDayResult.value = { date, ...result }
    }
  } catch (e) {
    toast('Network error', 'error')
  } finally {
    isCookingAll.value = false
  }
}

const closeCookAllResult = () => {
  cookAllDayResult.value = null
}
```

- [ ] **Step 3: Verify no console errors**

Start both servers:

```bash
cd backend && go run main.go &
cd frontend && npm run dev
```

Open http://localhost:5173, go to the Meal Plan tab, open browser DevTools console.
Expected: no errors, page loads normally.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/MealPlanner.vue
git commit -m "feat(frontend): add cook-all-day state and API functions"
```

---

### Task 6: Frontend — day card footer buttons

**Files:**
- Modify: `frontend/src/components/MealPlanner.vue`

- [ ] **Step 1: Replace the single Edit Day button**

Find this block in the template (line 96):

```html
          <!-- Edit Day Button -->
          <button @click.stop="openEdit(day.dateStr)"
                  class="mt-auto w-full text-center text-gray-300 hover:text-blue-500 hover:bg-blue-50 rounded-lg py-1.5 transition-all text-xs font-medium">Edit Day</button>
```

Replace with:

```html
          <!-- Day Actions -->
          <div class="mt-auto flex gap-1.5">
            <button
              v-if="getMealsForDay(day.dateStr).some(m => !m.is_cooked)"
              @click.stop="cookAllDay(day.dateStr)"
              :disabled="isCookingAll"
              class="flex-1 text-center text-orange-400 hover:text-orange-600 hover:bg-orange-50 rounded-lg py-1.5 transition-all text-xs font-medium disabled:opacity-40">
              🍳 Cook All
            </button>
            <button @click.stop="openEdit(day.dateStr)"
                    class="flex-1 text-center text-gray-300 hover:text-blue-500 hover:bg-blue-50 rounded-lg py-1.5 transition-all text-xs font-medium">Edit Day</button>
          </div>
```

- [ ] **Step 2: Verify in browser**

In the Meal Plan tab:
- A day with uncooked meals shows two buttons: "🍳 Cook All" and "Edit Day" side by side.
- A day where all meals are cooked (or no meals) shows only "Edit Day".
- "Edit Day" still opens the edit modal correctly.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/MealPlanner.vue
git commit -m "feat(frontend): add Cook All button to day card footer"
```

---

### Task 7: Frontend — confirm dialog

**Files:**
- Modify: `frontend/src/components/MealPlanner.vue`

- [ ] **Step 1: Add the confirm dialog block**

In the template, after the `ConfirmDialog` component block (after line 225), add:

```html
    <!-- Cook All Day Confirm Dialog -->
    <div v-if="confirmCookAllDayDate"
         class="fixed inset-0 bg-black/30 backdrop-blur-sm flex items-center justify-center z-50 p-4"
         @click.self="confirmCookAllDayDate = null">
      <div class="bg-white rounded-2xl shadow-2xl w-full max-w-sm overflow-hidden">
        <div class="p-6 border-b border-gray-100">
          <h3 class="font-bold text-xl text-gray-800">Cook all meals?</h3>
          <p class="text-sm text-gray-400 mt-0.5">{{ formatDateDisplay(confirmCookAllDayDate) }}</p>
        </div>
        <div class="p-6 space-y-2">
          <template v-for="type in ['Breakfast', 'Lunch', 'Dinner']" :key="type">
            <div v-for="meal in getMealsForDay(confirmCookAllDayDate).filter(m => m.meal_type === type && !m.is_cooked)"
                 :key="meal.id"
                 class="flex items-center gap-2 px-3 py-2 rounded-lg"
                 :class="mealTypeStyle(type).bg">
              <span>{{ mealTypeStyle(type).icon }}</span>
              <span class="text-sm font-medium text-gray-700">{{ meal.custom_name || meal.recipe_name || 'Unnamed' }}</span>
            </div>
          </template>
          <p class="text-xs text-amber-600 bg-amber-50 rounded-lg px-3 py-2 mt-2">
            ⚠ Some meals may fail if ingredients are low — we'll cook what we can
          </p>
        </div>
        <div class="flex border-t border-gray-100">
          <button @click="confirmCookAllDayDate = null"
                  class="flex-1 py-3.5 text-gray-500 hover:bg-gray-50 transition-colors font-semibold text-sm">
            Cancel
          </button>
          <button @click="confirmCookAllDayAction"
                  :disabled="isCookingAll"
                  class="flex-1 py-3.5 text-orange-500 hover:bg-orange-50 transition-colors font-bold text-sm border-l border-gray-100 disabled:opacity-40">
            {{ isCookingAll ? 'Cooking...' : 'Cook All 🍳' }}
          </button>
        </div>
      </div>
    </div>
```

- [ ] **Step 2: Verify in browser**

Click "🍳 Cook All" on a day with uncooked meals:
- Dialog appears with the day's uncooked meals listed (with meal type colors and icons).
- Amber warning at the bottom.
- "Cancel" dismisses. Clicking the backdrop dismisses.
- "Cook All 🍳" button is present.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/MealPlanner.vue
git commit -m "feat(frontend): add Cook All Day confirm dialog"
```

---

### Task 8: Frontend — result modal and E2E verification

**Files:**
- Modify: `frontend/src/components/MealPlanner.vue`

- [ ] **Step 1: Add the result modal block**

After the confirm dialog block, add:

```html
    <!-- Cook All Day Result Modal -->
    <div v-if="cookAllDayResult"
         class="fixed inset-0 bg-black/30 backdrop-blur-sm flex items-center justify-center z-50 p-4"
         @click.self="closeCookAllResult">
      <div class="bg-white rounded-2xl shadow-2xl w-full max-w-sm overflow-hidden">
        <div class="p-6 border-b border-gray-100">
          <h3 class="font-bold text-xl text-gray-800">Cooking done 🍳</h3>
          <p class="text-sm text-gray-400 mt-0.5">{{ formatDateDisplay(cookAllDayResult.date) }}</p>
        </div>
        <div class="p-6 space-y-4 max-h-[60vh] overflow-y-auto">
          <div v-if="cookAllDayResult.cooked.length > 0">
            <p class="text-xs font-bold text-green-700 uppercase tracking-wider mb-2">
              ✅ Cooked ({{ cookAllDayResult.cooked.length }})
            </p>
            <div class="space-y-1">
              <div v-for="meal in cookAllDayResult.cooked" :key="meal.id"
                   class="text-sm text-gray-700 bg-green-50 rounded-lg px-3 py-2">
                {{ meal.name }}
              </div>
            </div>
          </div>
          <div v-if="cookAllDayResult.failed.length > 0">
            <p class="text-xs font-bold text-red-600 uppercase tracking-wider mb-2">
              ❌ Couldn't Cook ({{ cookAllDayResult.failed.length }})
            </p>
            <div class="space-y-2">
              <div v-for="meal in cookAllDayResult.failed" :key="meal.id"
                   class="bg-red-50 rounded-lg px-3 py-2">
                <p class="text-sm font-semibold text-gray-700 mb-1">{{ meal.name }}</p>
                <p v-for="(m, i) in meal.missing" :key="i"
                   class="text-xs text-red-700">• {{ m }}</p>
              </div>
            </div>
          </div>
        </div>
        <div class="p-4 border-t border-gray-100">
          <button @click="closeCookAllResult"
                  class="w-full py-2.5 rounded-xl bg-gray-900 text-white font-semibold text-sm hover:bg-gray-800 transition-colors">
            OK
          </button>
        </div>
      </div>
    </div>
```

- [ ] **Step 2: Full E2E test**

With both dev servers running (backend on :8080, frontend on :5173):

1. Go to Meal Plan tab. Find a day with at least one uncooked meal that has a recipe.
2. Click "🍳 Cook All" → confirm dialog appears listing the uncooked meals.
3. Click "Cook All 🍳" → dialog closes, then result modal appears.
4. Verify: cooked meals appear in the green "✅ Cooked" section.
5. If any meal failed: verify it appears in the red "❌ Couldn't Cook" section with ingredient detail lines (e.g. "Rice (Need: 3, Have: 0)").
6. Click "OK" → modal closes.
7. Verify: the day card now shows cooked meals with 😋 icon and strikethrough text.
8. Verify: "🍳 Cook All" button disappears if all meals for that day are now cooked.
9. Verify: "Edit Day" still works and the edit modal opens correctly.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/MealPlanner.vue
git commit -m "feat(frontend): add Cook All Day result modal"
```
