// @ts-check
import { test, expect } from '@playwright/test'

const BASE_API = 'http://localhost:8080'

// Unique names so tests don't collide with real data
const INGREDIENT = 'E2E_Tofu'
const RECIPE = 'E2E_Recipe'

// Helpers ─────────────────────────────────────────────────────────────────────

/** Navigate to a tab by its desktop nav label */
async function goTab(page, label) {
  await page.getByRole('button', { name: label, exact: true }).first().click()
  await page.waitForLoadState('networkidle')
}

/** Today as YYYY-MM-DD in local time (same as the frontend) */
function todayStr() {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

// Cleanup helper – delete test data via API after all tests
test.afterAll(async ({ request }) => {
  // Fetch all recipes and delete E2E_Recipe
  const recipeRes = await request.get(`${BASE_API}/api/recipes`)
  const recipes = await recipeRes.json()
  for (const r of recipes || []) {
    if (r.name === RECIPE) {
      await request.post(`${BASE_API}/api/recipes/delete`, { data: { id: r.id } })
    }
  }

  // Fetch all inventory and delete E2E_Tofu
  const invRes = await request.get(`${BASE_API}/api/inventory`)
  const inventory = await invRes.json()
  for (const i of inventory || []) {
    if (i.name === INGREDIENT) {
      await request.post(`${BASE_API}/api/inventory/delete`, { data: { id: i.id } })
    }
  }

  // Delete E2E_Custom_Item from shopping list
  const shopRes = await request.get(`${BASE_API}/api/shopping-list`)
  const shopItems = await shopRes.json()
  for (const s of shopItems || []) {
    if (s.name === 'E2E_Custom_Item') {
      await request.post(`${BASE_API}/api/shopping-list/delete`, { data: { id: s.id } })
    }
  }
})

// ── Tests (run in order, share DB state) ─────────────────────────────────────
test.describe.serial('House Management E2E', () => {

  // ── 1. INVENTORY: Add ingredient ──────────────────────────────────────────
  test('Inventory – add new ingredient', async ({ page }) => {
    await page.goto('/')
    await goTab(page, 'Inventory')

    await page.getByRole('button', { name: 'Add Item' }).click()

    await page.fill('input[placeholder="e.g. Chicken, Soap..."]', INGREDIENT)
    await page.fill('input[placeholder="0"]', '0')

    await page.getByRole('button', { name: 'Add' }).click()

    await expect(page.locator(`td:has-text("${INGREDIENT}")`).first()).toBeVisible()
  })

  // ── 2. INVENTORY: Stock In with two batches ────────────────────────────────
  test('Inventory – stock in with two batches (FEFO setup)', async ({ page }) => {
    await page.goto('/')
    await goTab(page, 'Inventory')

    await page.getByRole('button', { name: 'Stock In' }).click()
    // Stock In modal has max-w-lg (Edit modal has max-w-md)
    const modal = page.locator('.max-w-lg')
    await expect(modal).toBeVisible()

    // Row 1: qty=1, expiry=2026-03-25 (earliest – should be consumed first)
    await modal.locator('select').first().selectOption({ label: INGREDIENT })
    await modal.locator('input[type="number"]').first().fill('1')
    await modal.locator('input[type="date"]').first().fill('2026-03-25')

    // Add second row
    await modal.getByText('+ Add row').click()
    await modal.locator('select').nth(1).selectOption({ label: INGREDIENT })
    await modal.locator('input[type="number"]').nth(1).fill('2')
    await modal.locator('input[type="date"]').nth(1).fill('2026-04-10')

    await modal.getByRole('button', { name: 'Confirm' }).click()
    await expect(modal).not.toBeVisible({ timeout: 5000 })

    // Stock should now be 3 (1+2)
    const row = page.locator('tr').filter({ hasText: INGREDIENT })
    await expect(row.locator('td').nth(2)).toHaveText('3')
  })

  // ── 3. INVENTORY: Edit modal shows correct batches ─────────────────────────
  test('Inventory – edit modal shows two batches', async ({ page }) => {
    await page.goto('/')
    await goTab(page, 'Inventory')

    const row = page.locator('tr').filter({ hasText: INGREDIENT })
    await row.getByRole('button', { name: 'Edit' }).click()

    const modal = page.locator('.fixed.inset-0').last()
    await expect(modal).toBeVisible()

    // Both batches should be listed (dates displayed as M/D by formatExpiry)
    await expect(modal.getByText('3/25')).toBeVisible()
    await expect(modal.getByText('4/10')).toBeVisible()

    await modal.getByRole('button', { name: 'Cancel' }).click()
  })

  // ── 4. RECIPES: Create recipe ─────────────────────────────────────────────
  test('Recipes – create recipe', async ({ page }) => {
    await page.goto('/')
    await goTab(page, 'Recipes')

    await page.fill('input[placeholder="What are you cooking?"]', RECIPE)
    await page.getByRole('button', { name: '+ Create' }).click()

    await expect(page.locator(`text=${RECIPE}`).first()).toBeVisible()
  })

  // ── 5. RECIPES: Add ingredient to recipe ──────────────────────────────────
  test('Recipes – add ingredient to recipe', async ({ page }) => {
    await page.goto('/')
    await goTab(page, 'Recipes')

    // Expand the E2E_Recipe card specifically
    const recipeCard = page.locator('.rounded-2xl').filter({ hasText: RECIPE })
    await recipeCard.locator('.cursor-pointer').filter({ hasText: 'Tap to add ingredients' }).click()
    await expect(page.locator('input[placeholder="e.g. Tofu, Chicken..."]')).toBeVisible()

    // Fill in the ingredient form
    await page.fill('input[placeholder="e.g. Tofu, Chicken..."]', INGREDIENT)
    await page.fill('input[placeholder="Qty"]', '1')
    await page.getByRole('button', { name: '+ Add Ingredient' }).click()

    // Ingredient tag should appear on the card
    await expect(page.locator('span.rounded-full').filter({ hasText: `${INGREDIENT} × 1` })).toBeVisible()
  })

  // ── 6. MEAL PLAN: Schedule recipe for today ────────────────────────────────
  test('Meal Plan – schedule recipe for today', async ({ page }) => {
    await page.goto('/')
    await goTab(page, 'Meal Plan')

    // Find today's column (has blue circle) and click its "Edit Day"
    const todayCol = page.locator('.text-center').filter({
      has: page.locator('.bg-blue-500'),
    })
    await todayCol.getByRole('button', { name: 'Edit Day' }).click()

    const modal = page.locator('.fixed.inset-0').last()
    await expect(modal).toBeVisible()

    // Select recipe in the Breakfast section (first select in modal)
    const selects = modal.locator('select')
    await selects.first().selectOption({ label: RECIPE })
    await modal.getByRole('button', { name: 'Add' }).first().click()

    // Recipe should appear as a pending new meal
    await expect(modal.locator(`text=${RECIPE}`).first()).toBeVisible()

    await modal.getByRole('button', { name: 'Save Changes' }).click()
    await expect(modal).not.toBeVisible()

    // Verify meal appears on the calendar
    await expect(page.locator(`text=${RECIPE}`).first()).toBeVisible()
  })

  // ── 7. MEAL PLAN: Cook meal (triggers FEFO) ────────────────────────────────
  test('Meal Plan – cook meal deducts earliest expiry batch (FEFO)', async ({ page }) => {
    await page.goto('/')
    await goTab(page, 'Meal Plan')

    // Open today's edit modal
    const todayCol = page.locator('.text-center').filter({
      has: page.locator('.bg-blue-500'),
    })
    await todayCol.getByRole('button', { name: 'Edit Day' }).click()

    const modal = page.locator('.fixed.inset-0').last()
    await expect(modal).toBeVisible()

    // Accept the confirm dialog automatically
    page.once('dialog', dialog => dialog.accept())
    await modal.locator('button[title="Cook this!"]').first().click()

    // Meal should now show as cooked (😋)
    await expect(modal.locator('text=😋').first()).toBeVisible()
    await modal.getByRole('button', { name: 'Cancel' }).click()
  })

  // ── 8. INVENTORY: Verify FEFO consumed earliest batch ─────────────────────
  test('Inventory – earliest batch consumed, later batch intact', async ({ page }) => {
    await page.goto('/')
    await goTab(page, 'Inventory')

    // Stock should now be 2 (3 - 1)
    const row = page.locator('tr').filter({ hasText: INGREDIENT })
    await expect(row.locator('td').nth(2)).toHaveText('2')

    // Open edit modal to inspect batches
    await row.getByRole('button', { name: 'Edit' }).click()
    const modal = page.locator('.fixed.inset-0').last()
    await expect(modal).toBeVisible()

    // Earliest batch (3/25, qty=1) should be GONE
    await expect(modal.getByText('3/25')).not.toBeVisible()

    // Later batch (4/10, qty=2) should still be there
    await expect(modal.getByText('4/10')).toBeVisible()

    await modal.getByRole('button', { name: 'Cancel' }).click()
  })

  // ── 9. SHOPPING LIST: Add and check off custom item ──────────────────────
  test('Shopping List – add custom item and check it off', async ({ page }) => {
    await page.goto('/')
    await goTab(page, 'Shopping')

    // Add a custom item
    await page.fill('input[placeholder="Add item to shopping list..."]', 'E2E_Custom_Item')
    await page.getByRole('button', { name: '+ Add' }).click()
    await expect(page.locator('text=E2E_Custom_Item').first()).toBeVisible()

    // Check off the custom item
    const item = page.locator('div.rounded-xl').filter({ hasText: 'E2E_Custom_Item' })
    await item.locator('input[type="checkbox"]').uncheck()
    await expect(item.locator('p').first()).toHaveClass(/line-through/)
  })
})
