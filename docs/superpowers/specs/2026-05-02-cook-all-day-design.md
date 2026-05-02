# Cook All Day — Design Spec

**Date:** 2026-05-02
**Branch:** feature/ux-overhaul

## Goal

Add a "Cook All Day" button to each day card in the meal plan calendar. Pressing it cooks all uncooked meals for that day at once, decrementing stock for all their ingredients.

## Decisions

| Question | Decision |
|---|---|
| Button placement | Two buttons at bottom of day card: `🍳 Cook All` + `Edit Day` |
| Confirmation | Dialog listing uncooked meals + generic low-stock warning |
| Stock failures | Cook what's possible; show result modal with cooked/failed breakdown |
| Result display | Modal: green cooked list + red failed list with missing ingredient details |
| Already-cooked meals | Silently skipped (backend queries `WHERE is_cooked = FALSE`) |
| Custom-name-only meals (no recipe, no ingredients) | Marked as cooked, stock step skipped |

## Backend

### New route

```
POST /api/mealplan/cook-day
```

Registered in `main.go`:
```go
mux.HandleFunc("POST /api/mealplan/cook-day", mealPlanHandler.CookAllDay)
```

### Request

```json
{ "date": "YYYY-MM-DD" }
```

### Handler logic (`CookAllDay` on `MealPlanHandler`)

1. Query all meals for the given date where `is_cooked = FALSE`.
2. For each meal — **one transaction per meal** (partial success is intentional):
   a. Mark `is_cooked = TRUE`.
   b. Load ingredients from `meal_plan_ingredients`. If none, backfill from `recipe_ingredients` (same fallback as `CookMeal`).
   c. If meal has no `recipe_id` and no `meal_plan_ingredients` — commit (mark cooked), skip stock step.
   d. Check stock sufficiency. If any ingredient is short — rollback this meal's transaction, record as failed with missing ingredient strings (format: `"Name (Need: X, Have: Y)"`).
   e. If stock is sufficient — apply FEFO batch decrements (same logic as `CookMeal`), recompute `current_stock`, commit.
3. Return HTTP 200 with combined result regardless of partial failures.

### Response

```json
{
  "cooked": [
    { "id": 1, "name": "Chicken Rice" }
  ],
  "failed": [
    { "id": 2, "name": "Pasta Bolognese", "missing": ["Ground Beef (Need: 3, Have: 1)", "Tomato Sauce (Need: 2, Have: 0)"] }
  ]
}
```

Both arrays are always present (empty array if none).

## Frontend (`MealPlanner.vue`)

### New state

```js
const confirmCookAllDayDate = ref(null)  // dateStr | null — controls confirm dialog
const cookAllDayResult = ref(null)       // { date, cooked: [], failed: [] } | null — controls result modal
const isCookingAll = ref(false)
```

### New functions

- `cookAllDay(dateStr)` — sets `confirmCookAllDayDate` to open the confirm dialog.
- `confirmCookAllDayAction()` — sets `isCookingAll = true`, calls `POST /api/mealplan/cook-day`, stores response in `cookAllDayResult`, calls `fetchMealPlan()`.
- `closeCookAllResult()` — clears `cookAllDayResult`.

### Day card footer

Replace the single `Edit Day` button with:

```
[uncooked meals exist]   →  [ 🍳 Cook All ]  [ Edit Day ]
[all cooked / no meals]  →             [ Edit Day ]
```

Visibility condition: `getMealsForDay(dateStr).some(m => !m.is_cooked)`

### Confirm dialog (inline `v-if` block, not a new component)

Shown when `confirmCookAllDayDate !== null`.

Content:
- Title: "Cook all meals?"
- Subtitle: the formatted date
- List of uncooked meals for that day, grouped by meal type with icons (🌅 / ☀️ / 🌙)
- Generic warning at bottom: *"Some meals may fail if ingredients are low — we'll cook what we can"*
- Buttons: `Cancel` | `Cook All 🍳` (disabled + spinner while `isCookingAll`)

### Result modal (inline `v-if` block)

Shown when `cookAllDayResult !== null`.

Content:
- Header: "Cooking done 🍳" + date
- Green section (if `cooked.length > 0`): "✅ COOKED (N)" + meal name list
- Red section (if `failed.length > 0`): "❌ COULDN'T COOK (N)" + per-meal block with missing ingredient strings
- If `cooked.length === 0` and `failed.length === 0`: "Nothing to cook" (all meals were already cooked or skipped)
- Single "OK" button → calls `closeCookAllResult()`

## Edge cases

| Case | Behavior |
|---|---|
| No uncooked meals for the day | "Cook All" button not shown |
| All meals fail stock check | Result modal shows only red section |
| All meals succeed | Result modal shows only green section |
| Day has only custom-name meals (no ingredients) | All marked cooked, green section shows them, no stock decremented |
| User clicks "Cook All" while another is in-flight | Button disabled while `isCookingAll` |
