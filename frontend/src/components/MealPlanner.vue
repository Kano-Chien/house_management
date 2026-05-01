<template>
  <div class="p-6">
    <h2 class="text-2xl font-bold mb-2 text-gray-800">📅 Meal Planner</h2>

    <!-- Week Navigation -->
    <div class="flex justify-between items-center mb-5">
      <button @click="prevWeek" class="bg-white border border-gray-200 px-4 py-2 rounded-xl hover:bg-gray-50 transition-colors shadow-sm font-medium text-gray-600">← Prev</button>
      <span class="font-semibold text-gray-600">{{ weekLabel }}</span>
      <button @click="nextWeek" class="bg-white border border-gray-200 px-4 py-2 rounded-xl hover:bg-gray-50 transition-colors shadow-sm font-medium text-gray-600">Next →</button>
    </div>

    <!-- Weekly Calendar Grid -->
    <div class="grid grid-cols-1 md:grid-cols-7 gap-3 pb-24 md:pb-0">
      <!-- Day Headers -->
      <div v-for="day in weekDays" :key="day.dateStr"
           class="text-center">
        <div class="text-xs uppercase tracking-wider font-semibold text-gray-400 mb-1">{{ day.dayName }}</div>
        <div :class="[day.isToday ? 'bg-blue-500 text-white' : 'bg-gray-100 text-gray-600']"
             class="rounded-full w-8 h-8 flex items-center justify-center mx-auto text-sm font-bold mb-2">
          {{ day.dayNum }}
        </div>

        <!-- Day Card -->
        <div class="bg-white rounded-xl border border-gray-100 shadow-sm min-h-[140px] p-2 flex flex-col gap-2">
          <!-- Iterate over meal types -->
          <div v-for="type in ['Breakfast', 'Lunch', 'Dinner']" :key="type">
            <!-- Show block only if there are meals for this type -->
            <div v-if="getMealsForDay(day.dateStr).filter(m => m.meal_type === type).length > 0"
                 class="rounded-lg p-2 transition-all border mb-1"
                 :class="mealTypeStyle(type).bg">
              <!-- Header -->
              <div class="text-[10px] uppercase tracking-wider font-bold mb-1"
                   :class="mealTypeStyle(type).text">
                {{ mealTypeStyle(type).icon }} {{ type }}
              </div>
              <!-- Meals List -->
                  <div class="space-y-1">
                <div v-for="meal in getMealsForDay(day.dateStr).filter(m => m.meal_type === type)" :key="meal.id"
                     class="group relative pl-1.5 pr-1 py-1 hover:bg-white/60 transition-colors rounded-md min-h-[1.75rem]">

                  <!-- Meal Name -->
                  <div class="text-xs font-medium leading-tight flex items-start gap-1"
                       :class="meal.is_cooked ? 'text-gray-400 font-normal line-through opacity-70' : 'text-gray-700'"
                       :title="meal.custom_name || meal.recipe_name">
                    <span v-if="meal.is_cooked" class="select-none text-xs pt-0.5" title="Yummy!">😋</span>
                    <span>{{ meal.custom_name || meal.recipe_name || 'Unnamed' }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Edit Day Button (replaces individual Add button) -->
          <button @click.stop="openEdit(day.dateStr)"
                  class="mt-auto w-full text-center text-gray-300 hover:text-blue-500 hover:bg-blue-50 rounded-lg py-1 transition-all text-sm font-medium">Edit Day</button>
        </div>
      </div>
    </div>

    <!-- Edit Day Modal -->
    <div v-if="showModal" class="fixed inset-0 bg-black/30 backdrop-blur-sm flex items-center justify-center z-50 p-4" @click.self="closeModal">
      <div class="bg-white rounded-2xl shadow-2xl w-full max-w-md flex flex-col max-h-[90vh]">

        <!-- Header -->
        <div class="p-6 border-b border-gray-100">
          <h3 class="font-bold text-xl text-gray-800">Edit Meals</h3>
          <p class="text-sm text-gray-400">{{ formatDateDisplay(modalDate) }}</p>
        </div>

        <!-- Scrollable Content -->
        <div class="p-6 overflow-y-auto flex-1 space-y-6">
          <div v-for="type in ['Breakfast', 'Lunch', 'Dinner']" :key="type">
            <h4 class="text-xs font-bold uppercase tracking-wider text-gray-400 mb-2 flex items-center gap-2">
              {{ mealTypeStyle(type).icon }} {{ type }}
            </h4>

            <!-- Existing Meals (that are not marked for removal) -->
            <div class="space-y-2 mb-2">
              <div v-for="meal in getMealsForEditor(type)" :key="'prod-'+meal.id"
                   class="flex items-center justify-between bg-gray-50 p-2 rounded-lg group">
                <div class="flex items-center gap-2 flex-1 min-w-0">
                   <!-- Cook Button/Status -->
                   <button v-if="!meal.is_cooked" @click.stop="cookMeal(meal)"
                           class="text-amber-500 hover:text-amber-600 hover:bg-amber-50 p-1 rounded transition-colors text-lg leading-none"
                           title="Cook this!">🍳</button>
                   <span v-else class="text-lg leading-none select-none" title="Cooked">😋</span>

                   <span class="text-sm font-medium text-gray-700 truncate" :class="{'line-through text-gray-400': meal.is_cooked}">
                     {{ getUpdatedName(meal) }}
                   </span>
                </div>
                <!-- Edit -->
                <button v-if="!meal.is_cooked && meal.recipe_id" @click="editExistingMeal(meal)" class="text-gray-300 hover:text-blue-400 p-1 transition-colors text-sm leading-none" title="Edit ingredients">✏️</button>
                <!-- Delete -->
                <button @click="markForRemoval(meal.id)" class="text-gray-300 hover:text-red-400 p-1 ml-1 transition-colors text-lg leading-none">🗑️</button>
              </div>

              <!-- Newly Added Meals (Pending Save) -->
              <div v-for="meal in newMeals.filter(m => m.meal_type === type)" :key="meal._key"
                   class="flex items-center justify-between bg-blue-50 border border-blue-100 p-2 rounded-lg">
                <span class="text-sm font-medium text-blue-700 truncate flex-1">{{ meal.custom_name || getRecipeName(meal.recipe_id) }}</span>
                <button @click="removeNewMeal(meal)" class="text-blue-400 hover:text-blue-600 p-1">✕</button>
              </div>
            </div>

            <!-- Add New Meal Controls -->
            <div class="space-y-2">
              <div v-if="!editingMealId || !showIngredientEditor[type]" class="flex gap-2">
                <select v-model="selectedRecipes[type]" @change="onRecipeSelected(type)" class="flex-1 border border-gray-200 rounded-lg text-sm p-1.5 focus:outline-none focus:ring-2 focus:ring-blue-100">
                  <option value="" disabled>Add recipe...</option>
                  <option v-for="r in recipes" :key="r.id" :value="r.id">{{ r.name }}</option>
                </select>
              </div>

              <!-- Ingredient Editor (shown after recipe selection) -->
              <div v-if="showIngredientEditor[type]" class="bg-gray-50 rounded-lg p-3 space-y-3 border border-gray-200">
                <!-- Variant Name -->
                <div>
                  <label class="text-xs font-medium text-gray-500 mb-1 block">Variant name (optional)</label>
                  <input v-model="variantNames[type]" type="text" :placeholder="getRecipeName(selectedRecipes[type]) + ' variant...'"
                         class="w-full border border-gray-200 rounded-lg text-sm p-1.5 focus:outline-none focus:ring-2 focus:ring-blue-100" />
                </div>

                <!-- Ingredient List -->
                <div>
                  <label class="text-xs font-medium text-gray-500 mb-1 block">Ingredients</label>
                  <div class="space-y-1.5">
                    <div v-for="(ing, idx) in editingIngredients[type]" :key="ing.ingredient_id"
                         class="flex items-center gap-2 bg-white rounded-md p-1.5 border border-gray-100">
                      <span class="text-sm text-gray-700 flex-1 truncate">{{ ing.name }}</span>
                      <input v-model.number="ing.quantity" type="number" min="1"
                             class="w-16 border border-gray-200 rounded text-sm p-1 text-center focus:outline-none focus:ring-2 focus:ring-blue-100" />
                      <button @click="editingIngredients[type].splice(idx, 1)" class="text-gray-300 hover:text-red-400 text-sm px-1">✕</button>
                    </div>
                  </div>
                </div>

                <!-- Add Ingredient -->
                <div>
                  <select v-model="addIngredientSelection[type]" @change="addIngredientToEditor(type)" class="w-full border border-gray-200 rounded-lg text-sm p-1.5 focus:outline-none focus:ring-2 focus:ring-blue-100">
                    <option value="" disabled>Add ingredient...</option>
                    <option v-for="inv in availableIngredients(type)" :key="inv.id" :value="inv.id">{{ inv.name }}</option>
                  </select>
                </div>

                <!-- Save / Cancel -->
                <div class="flex gap-2 pt-1">
                  <button @click="cancelIngredientEditor(type)"
                          class="flex-1 py-1.5 rounded-lg bg-white border border-gray-200 text-gray-500 hover:bg-gray-50 text-sm font-medium">Cancel</button>
                  <button v-if="editingMealId" @click="saveEditingMeal(type)"
                          class="flex-1 py-1.5 rounded-lg bg-green-500 text-white hover:bg-green-600 text-sm font-medium transition-colors">Save changes</button>
                  <button v-else @click="addNewMeal(type)"
                          class="flex-1 py-1.5 rounded-lg bg-blue-500 text-white hover:bg-blue-600 text-sm font-medium transition-colors">Add to meal</button>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Footer -->
        <div class="p-4 border-t border-gray-100 bg-gray-50 rounded-b-2xl flex gap-3">
          <button @click="closeModal"
                  class="flex-1 py-2.5 rounded-xl bg-white border border-gray-200 text-gray-600 hover:bg-gray-50 transition-colors font-semibold text-sm">Cancel</button>
          <button @click="saveChanges"
                  class="flex-1 py-2.5 rounded-xl bg-gradient-to-r from-blue-500 to-indigo-600 text-white hover:shadow-lg transition-all font-semibold text-sm shadow-md disabled:opacity-50 disabled:cursor-wait"
                  :disabled="isSaving">
            {{ isSaving ? 'Saving...' : 'Save Changes' }}
          </button>
        </div>
      </div>
    </div>

    <ConfirmDialog
      v-if="confirmCookMeal"
      title="Mark as cooked?"
      :message="`Mark &quot;${confirmCookMeal.recipe_name || confirmCookMeal.custom_name}&quot; as cooked?`"
      confirmLabel="Cook it 🍳"
      confirmVariant="primary"
      @confirm="confirmCookMealItem"
      @cancel="confirmCookMeal = null"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import ConfirmDialog from './ui/ConfirmDialog.vue'
import { useToast } from '../composables/useToast.js'

const { toast } = useToast()
const confirmCookMeal = ref(null)

const mealPlan = ref([])
const recipes = ref([])
const inventory = ref([])
const weekOffset = ref(0)
const showModal = ref(false)
const modalDate = ref('')
const isSaving = ref(false)

// Editor State
const removedMealIds = ref(new Set())
const newMeals = ref([]) // { meal_type, recipe_id, custom_name, ingredients: [] }
const selectedRecipes = ref({ Breakfast: '', Lunch: '', Dinner: '' })
const variantNames = ref({ Breakfast: '', Lunch: '', Dinner: '' })
const editingIngredients = ref({ Breakfast: [], Lunch: [], Dinner: [] })
const showIngredientEditor = ref({ Breakfast: false, Lunch: false, Dinner: false })
const addIngredientSelection = ref({ Breakfast: '', Lunch: '', Dinner: '' })
const editingMealId = ref(null) // meal_plan id being edited (null = adding new)
const updatedMeals = ref([]) // { id, custom_name, ingredients: [] }

const getMonday = (offset) => {
  const d = new Date()
  const day = d.getDay()
  const diff = d.getDate() - day + (day === 0 ? -6 : 1) + (offset * 7)
  const monday = new Date(d.setDate(diff))
  monday.setHours(0, 0, 0, 0)
  return monday
}

// Format date as YYYY-MM-DD using local timezone
const formatDate = (d) => {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

const weekDays = computed(() => {
  const monday = getMonday(weekOffset.value)
  const days = []
  const dayNames = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']
  const today = formatDate(new Date())

  for (let i = 0; i < 7; i++) {
    const d = new Date(monday)
    d.setDate(monday.getDate() + i)
    const dateStr = formatDate(d)
    days.push({
      dayName: dayNames[i],
      dayNum: d.getDate(),
      dateStr,
      isToday: dateStr === today,
    })
  }
  return days
})

const weekLabel = computed(() => {
  const days = weekDays.value
  const start = new Date(days[0].dateStr)
  const end = new Date(days[6].dateStr)
  const opts = { month: 'short', day: 'numeric' }
  return `${start.toLocaleDateString('en-US', opts)} – ${end.toLocaleDateString('en-US', opts)}, ${end.getFullYear()}`
})

const getMealsForDay = (dateStr) => {
  return mealPlan.value
    .filter(m => {
      // Parse to local date string to avoid UTC offset issues
      const d = new Date(m.date)
      const mDate = `${d.getFullYear()}-${String(d.getMonth()+1).padStart(2,'0')}-${String(d.getDate()).padStart(2,'0')}`
      return mDate === dateStr
    })
    .sort((a, b) => {
      const order = { Breakfast: 0, Lunch: 1, Dinner: 2 }
      return (order[a.meal_type] ?? 9) - (order[b.meal_type] ?? 9)
    })
}

// Editor Helper: Get meals for the modal that are NOT removed
const getMealsForEditor = (type) => {
  return getMealsForDay(modalDate.value)
    .filter(m => m.meal_type === type && !removedMealIds.value.has(m.id))
}

const getRecipeName = (id) => recipes.value.find(r => r.id === id)?.name || 'Unknown'

const getUpdatedName = (meal) => {
  const updated = updatedMeals.value.find(m => m.id === meal.id)
  if (updated && updated.custom_name) return updated.custom_name
  return meal.custom_name || meal.recipe_name || 'Unnamed'
}

const fetchRecipes = async () => {
  try {
    const res = await fetch('/api/recipes')
    if (res.ok) recipes.value = (await res.json()) || []
  } catch (e) { toast('Network error', 'error') }
}

const fetchMealPlan = async () => {
  try {
    const res = await fetch('/api/mealplan')
    if (res.ok) mealPlan.value = (await res.json()) || []
  } catch (e) { toast('Network error', 'error') }
}

const fetchInventory = async () => {
  try {
    const res = await fetch('/api/inventory')
    if (res.ok) inventory.value = (await res.json()) || []
  } catch (e) { toast('Network error', 'error') }
}

const mealTypeStyle = (type) => {
  switch (type) {
    case 'Breakfast': return { bg: 'bg-emerald-50 border border-emerald-100', text: 'text-emerald-500', icon: '🌅' }
    case 'Lunch': return { bg: 'bg-amber-50 border border-amber-100', text: 'text-amber-500', icon: '☀️' }
    case 'Dinner': return { bg: 'bg-indigo-50 border border-indigo-100', text: 'text-indigo-500', icon: '🌙' }
    default: return { bg: 'bg-gray-50 border border-gray-100', text: 'text-gray-500', icon: '🍽️' }
  }
}

const openEdit = (dateStr) => {
  modalDate.value = dateStr
  removedMealIds.value = new Set()
  newMeals.value = []
  updatedMeals.value = []
  editingMealId.value = null
  selectedRecipes.value = { Breakfast: '', Lunch: '', Dinner: '' }
  variantNames.value = { Breakfast: '', Lunch: '', Dinner: '' }
  editingIngredients.value = { Breakfast: [], Lunch: [], Dinner: [] }
  showIngredientEditor.value = { Breakfast: false, Lunch: false, Dinner: false }
  addIngredientSelection.value = { Breakfast: '', Lunch: '', Dinner: '' }
  showModal.value = true
}

const closeModal = () => showModal.value = false

const markForRemoval = (id) => removedMealIds.value.add(id)

const onRecipeSelected = async (type) => {
  const recipeId = selectedRecipes.value[type]
  if (!recipeId) return

  try {
    const res = await fetch(`/api/recipes/ingredients?recipe_id=${recipeId}`)
    if (res.ok) {
      const ingredients = (await res.json()) || []
      editingIngredients.value[type] = ingredients.map(ing => ({
        ingredient_id: ing.ingredient_id,
        name: ing.name,
        quantity: ing.quantity,
      }))
    }
  } catch (e) { toast('Network error', 'error') }

  variantNames.value[type] = ''
  showIngredientEditor.value[type] = true
}

const availableIngredients = (type) => {
  const usedIds = new Set(editingIngredients.value[type].map(i => i.ingredient_id))
  return inventory.value.filter(i => !usedIds.has(i.id))
}

const addIngredientToEditor = (type) => {
  const id = addIngredientSelection.value[type]
  if (!id) return
  const inv = inventory.value.find(i => i.id === id)
  if (!inv) return
  editingIngredients.value[type].push({
    ingredient_id: inv.id,
    name: inv.name,
    quantity: 1,
  })
  addIngredientSelection.value[type] = ''
}

const editExistingMeal = async (meal) => {
  if (meal.is_cooked) return
  const type = meal.meal_type

  const draftMeal = updatedMeals.value.find(m => m.id === meal.id)

  if (draftMeal) {
    editingIngredients.value[type] = draftMeal.ingredients.map(ing => ({
      ingredient_id: ing.ingredient_id,
      name: ing.name,
      quantity: ing.quantity,
    }))
  } else {
    try {
      const res = await fetch(`/api/mealplan/ingredients?meal_plan_id=${meal.id}`)
      if (res.ok) {
        const ingredients = (await res.json()) || []
        editingIngredients.value[type] = ingredients.map(ing => ({
          ingredient_id: ing.ingredient_id,
          name: ing.name,
          quantity: ing.quantity,
        }))
      }
    } catch (e) { toast('Network error', 'error') }
  }

  editingMealId.value = meal.id
  selectedRecipes.value[type] = meal.recipe_id || ''
  
  variantNames.value[type] = draftMeal ? draftMeal.custom_name : (meal.custom_name || '')
  
  showIngredientEditor.value[type] = true
}

const saveEditingMeal = (type) => {
  const mealId = editingMealId.value
  if (!mealId) return

  const customName = variantNames.value[type].trim()
  
  const draftIngredients = editingIngredients.value[type]
    .filter(ing => ing.quantity > 0)
    .map(ing => ({
      ingredient_id: ing.ingredient_id,
      name: ing.name,
      quantity: ing.quantity,
    }))

  updatedMeals.value = updatedMeals.value.filter(m => m.id !== mealId)
  updatedMeals.value.push({ 
    id: mealId, 
    custom_name: customName, 
    ingredients: draftIngredients 
  })
  
  cancelIngredientEditor(type)
}

const cancelIngredientEditor = (type) => {
  showIngredientEditor.value[type] = false
  selectedRecipes.value[type] = ''
  variantNames.value[type] = ''
  editingIngredients.value[type] = []
  addIngredientSelection.value[type] = ''
  editingMealId.value = null
}

const addNewMeal = (type) => {
  const recipeId = selectedRecipes.value[type]
  if (!recipeId) return

  const customName = variantNames.value[type].trim()
  const ingredients = editingIngredients.value[type]
    .filter(ing => ing.quantity > 0)
    .map(ing => ({ ingredient_id: ing.ingredient_id, quantity: ing.quantity }))

  newMeals.value.push({
    _key: `new-meal-${Date.now()}-${Math.random()}`,
    meal_type: type,
    recipe_id: recipeId,
    custom_name: customName,
    ingredients
  })

  // Reset editor for this type
  cancelIngredientEditor(type)
}

const removeNewMeal = (mealObj) => {
  const idx = newMeals.value.indexOf(mealObj)
  if (idx > -1) newMeals.value.splice(idx, 1)
}

const saveChanges = async () => {
  isSaving.value = true
  try {
    const promises = []

    // Delete removed meals
    for (const id of removedMealIds.value) {
      promises.push(fetch('/api/mealplan/delete', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id })
      }))
    }

    // Update edited meals
    for (const meal of updatedMeals.value) {
      promises.push(fetch('/api/mealplan/update', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          id: meal.id,
          custom_name: meal.custom_name,
          ingredients: meal.ingredients
        })
      }))
    }

    // Add new meals
    for (const meal of newMeals.value) {
      promises.push(fetch('/api/mealplan', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          date: modalDate.value,
          meal_type: meal.meal_type,
          recipe_id: parseInt(meal.recipe_id),
          custom_name: meal.custom_name,
          ingredients: meal.ingredients
        })
      }))
    }

    await Promise.all(promises)
    await fetchMealPlan()
    closeModal()
  } catch (e) {
    toast('Failed to save changes', 'error')
  } finally {
    isSaving.value = false
  }
}

const cookMeal = (meal) => {
  confirmCookMeal.value = meal
}
const confirmCookMealItem = async () => {
  const meal = confirmCookMeal.value
  confirmCookMeal.value = null
  try {
    const res = await fetch('/api/mealplan/cook', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id: meal.id })
    })
    if (res.ok) {
      await fetchMealPlan()
      toast('Marked as cooked! 🍳')
    } else {
      toast('Failed to mark as cooked', 'error')
    }
  } catch (e) {
    toast('Failed to mark as cooked', 'error')
  }
}

const prevWeek = () => weekOffset.value--
const nextWeek = () => weekOffset.value++

const formatDateDisplay = (dateStr) => {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleDateString('en-US', { weekday: 'long', month: 'long', day: 'numeric' })
}

onMounted(() => {
  fetchRecipes()
  fetchMealPlan()
  fetchInventory()
})
</script>
