<template>
  <div class="p-6">
    <h2 class="text-2xl font-bold mb-6 text-gray-800">🍳 Recipes</h2>

    <!-- Create Recipe -->
    <div class="mb-8 p-5 bg-white rounded-2xl shadow-lg border border-gray-100">
      <div class="flex gap-3">
        <input v-model="newRecipeName" placeholder="What are you cooking?"
          @keyup.enter="addRecipe"
          class="border-2 border-gray-200 p-3 rounded-xl flex-1 text-base focus:ring-2 focus:ring-emerald-400 focus:border-emerald-400 focus:outline-none transition-all placeholder-gray-300" />
        <button @click="addRecipe"
          class="bg-gradient-to-r from-emerald-500 to-green-500 text-white px-6 py-3 rounded-xl hover:from-emerald-600 hover:to-green-600 transition-all font-semibold shadow-md hover:shadow-lg active:scale-95">
          + Create
        </button>
      </div>
    </div>

    <!-- Empty State -->
    <div v-if="recipes.length === 0" class="text-center py-16">
      <div class="text-6xl mb-4">📖</div>
      <p class="text-gray-400 text-lg">No recipes yet. Create your first one!</p>
    </div>

    <!-- Recipe Cards Grid -->
    <div class="grid gap-6 md:grid-cols-2">
      <div v-for="(recipe, idx) in recipes" :key="recipe.id"
           class="group rounded-2xl shadow-lg overflow-hidden border border-gray-100 transition-all duration-300 hover:shadow-xl hover:-translate-y-1">

        <!-- Card Header -->
        <div :class="cardColors[idx % cardColors.length]"
             class="p-5 relative overflow-hidden">
          <!-- Decorative circles -->
          <div class="absolute -top-6 -right-6 w-24 h-24 bg-white/10 rounded-full"></div>
          <div class="absolute -bottom-8 -left-4 w-20 h-20 bg-white/5 rounded-full"></div>
          <div class="flex justify-between items-center relative z-10">
            <div class="flex-1 min-w-0 mr-3">
              <!-- Editable Recipe Name -->
              <div v-if="editingNameId === recipe.id" class="flex gap-2 items-center">
                <input v-model="editNameValue" ref="nameEditInput"
                  @keyup.enter="saveRecipeName(recipe)"
                  @keyup.escape="editingNameId = null"
                  @blur="saveRecipeName(recipe)"
                  class="bg-white/20 backdrop-blur-sm text-white font-bold text-xl border border-white/30 rounded-lg px-2 py-1 w-full focus:outline-none focus:ring-2 focus:ring-white/50 placeholder-white/50" />
              </div>
              <div v-else class="cursor-pointer" @click="startEditName(recipe)">
                <h3 class="font-bold text-xl text-white tracking-wide flex items-center gap-2">
                  {{ recipe.name }}
                  <span class="opacity-0 group-hover:opacity-60 text-white/80 text-xs transition-opacity">✏️</span>
                </h3>
              </div>
              <p class="text-white/60 text-xs mt-1 cursor-pointer" @click="toggleRecipe(recipe.id)">
                {{ (recipe._ingredients || []).length }} ingredients · ${{ recipeTotalCost(recipe) }}
              </p>
            </div>
            <div class="flex gap-2 items-center flex-shrink-0">
              <span class="bg-white/20 backdrop-blur-sm text-white w-7 h-7 rounded-full flex items-center justify-center text-xs cursor-pointer"
                @click="toggleRecipe(recipe.id)">
                {{ expandedId === recipe.id ? '▲' : '▼' }}
              </span>
              <button @click.stop="deleteRecipe(recipe)"
                class="bg-white/10 backdrop-blur-sm hover:bg-red-500/80 text-white w-7 h-7 rounded-full flex items-center justify-center text-xs transition-all">✕</button>
            </div>
          </div>
        </div>

        <!-- Card Body -->
        <div class="bg-white">
          <!-- Ingredient Tags (always visible) -->
          <div class="px-5 pt-4 pb-1 flex flex-wrap gap-1.5 cursor-pointer" @click="toggleRecipe(recipe.id)">
            <span v-for="ing in sortedIngs(recipe._ingredients)" :key="ing.ingredient_id"
                  :class="ing.is_tracked === false
                    ? 'bg-gradient-to-r from-amber-50 to-orange-50 text-amber-700 border-amber-200'
                    : 'bg-gradient-to-r from-gray-50 to-gray-100 text-gray-600 border-gray-200'"
                  class="px-3 py-1 rounded-full text-xs font-medium border shadow-sm">
              {{ ing.name }} × {{ ing.quantity }}
            </span>
            <span v-if="!recipe._ingredients || recipe._ingredients.length === 0" class="text-gray-300 text-sm italic">
              Tap to add ingredients ↓
            </span>
          </div>


          <!-- Expanded Section -->
          <div v-if="expandedId === recipe.id" class="border-t border-gray-100 px-5 py-4">
            <!-- Ingredient List -->
            <div v-if="recipeIngredients.length > 0" class="mb-4 space-y-1">
              <div v-for="ing in sortedIngs(recipeIngredients)" :key="ing.ingredient_id"
                   class="flex justify-between items-center py-2 px-3 rounded-lg hover:bg-gray-50 transition-colors group/item">
                <div class="flex items-center gap-2">
                  <span :class="ing.is_tracked === false ? 'bg-amber-400' : 'bg-emerald-400'" class="w-2 h-2 rounded-full"></span>
                  <span class="text-sm text-gray-700 font-medium">{{ ing.name }}</span>

                  <!-- Editable Quantity -->
                  <span v-if="editingQty && editingQty.ingredient_id === ing.ingredient_id" class="flex items-center gap-1">
                    <span class="text-sm text-gray-400">×</span>
                    <input v-model.number="editQtyValue" type="number" step="1" min="1"
                      ref="qtyEditInput"
                      @keyup.enter="saveIngredientQty(recipe.id, ing)"
                      @keyup.escape="editingQty = null"
                      @blur="saveIngredientQty(recipe.id, ing)"
                      class="border border-blue-300 rounded px-1.5 py-0.5 w-14 text-sm text-center focus:ring-2 focus:ring-blue-400 focus:outline-none" />
                  </span>
                  <span v-else @click="startEditQty(ing)"
                    class="text-sm text-gray-400 cursor-pointer hover:text-blue-500 hover:bg-blue-50 px-1.5 py-0.5 rounded transition-all">
                    × {{ ing.quantity }}
                  </span>

                  <span v-if="ing.price" class="text-xs text-emerald-500 font-medium">${{ (ing.price * ing.quantity).toFixed(0) }}</span>
                  <span v-if="ing.unit" class="text-xs text-gray-300">{{ ing.unit }}</span>
                </div>
                <button @click="removeIngredient(recipe.id, ing.ingredient_id)"
                  class="opacity-0 group-hover/item:opacity-100 text-red-400 hover:text-red-600 transition-all text-lg leading-none">×</button>
              </div>
            </div>

            <!-- Add Ingredient Form -->
            <div class="p-4 bg-gradient-to-br from-gray-50 to-blue-50/30 rounded-xl border border-gray-100">
              <p class="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-3">Add ingredient</p>
              <div class="flex gap-2 mb-3">
                <input v-model="addIngForm.new_name" placeholder="e.g. Tofu, Chicken..."
                  @keyup.enter="addIngredientToRecipe(recipe.id)"
                  class="border-2 border-gray-200 p-2 rounded-lg text-sm flex-1 focus:ring-2 focus:ring-blue-400 focus:border-blue-400 focus:outline-none transition-all placeholder-gray-300" />
                <input v-model.number="addIngForm.quantity" type="number" step="1" min="1" placeholder="Qty"
                  @keyup.enter="addIngredientToRecipe(recipe.id)"
                  class="border-2 border-gray-200 p-2 rounded-lg text-sm w-16 text-center focus:ring-2 focus:ring-blue-400 focus:border-blue-400 focus:outline-none transition-all" />
              </div>
              <div class="flex items-center gap-2 mb-3">
                <input type="checkbox" v-model="addIngForm.is_tracked" :id="'track-'+recipe.id" class="rounded text-blue-500 focus:ring-blue-400 w-4 h-4 cursor-pointer">
                <label :for="'track-'+recipe.id" class="text-xs text-gray-500 font-medium cursor-pointer select-none">Track in Inventory</label>
              </div>
              <button @click="addIngredientToRecipe(recipe.id)"
                class="bg-gradient-to-r from-blue-500 to-indigo-500 text-white px-4 py-2 rounded-lg hover:from-blue-600 hover:to-indigo-600 transition-all text-sm font-semibold w-full shadow-md hover:shadow-lg active:scale-[0.98]">
                + Add Ingredient
              </button>
            </div>

            <!-- Recipe Steps (Notion-like WYSIWYG) -->
            <div class="mt-4 p-4 bg-blue-50/50 rounded-xl border border-blue-100">
              <p class="text-xs font-semibold text-blue-600 uppercase tracking-wider mb-2">📝 Recipe Steps</p>
              <StepEditor v-model="recipe.notes" @blur="saveNotes(recipe)" />
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue'
import StepEditor from './StepEditor.vue'

const recipes = ref([])
const newRecipeName = ref('')

// Sort tracked ingredients first, untracked last
const sortedIngs = (ings) => {
  if (!ings || !ings.length) return []
  return [...ings].sort((a, b) => {
    const aTracked = a.is_tracked !== false ? 0 : 1
    const bTracked = b.is_tracked !== false ? 0 : 1
    return aTracked - bTracked
  })
}
const expandedId = ref(null)
const recipeIngredients = ref([])
const addIngForm = ref({ new_name: '', quantity: 1, is_tracked: true })

// Editing state
const editingNameId = ref(null)
const editNameValue = ref('')
const editingQty = ref(null)
const editQtyValue = ref(1)
const nameEditInput = ref(null)
const qtyEditInput = ref(null)

const cardColors = [
  'bg-gradient-to-br from-blue-500 via-blue-600 to-indigo-600',
  'bg-gradient-to-br from-violet-500 via-purple-500 to-fuchsia-500',
  'bg-gradient-to-br from-amber-500 via-orange-500 to-red-400',
  'bg-gradient-to-br from-emerald-500 via-teal-500 to-cyan-500',
  'bg-gradient-to-br from-rose-500 via-pink-500 to-fuchsia-400',
  'bg-gradient-to-br from-sky-500 via-blue-500 to-indigo-500',
]

const recipeTotalCost = (recipe) => {
  const ings = recipe._ingredients || []
  return ings.reduce((sum, ing) => sum + (ing.price || 0) * ing.quantity, 0).toFixed(0)
}

// ---- Recipe Name Editing ----
const startEditName = async (recipe) => {
  editingNameId.value = recipe.id
  editNameValue.value = recipe.name
  await nextTick()
  const inputs = nameEditInput.value
  if (Array.isArray(inputs) && inputs.length) inputs[0].focus()
  else if (inputs) inputs.focus()
}

const saveRecipeName = async (recipe) => {
  if (!editNameValue.value || editNameValue.value === recipe.name) {
    editingNameId.value = null
    return
  }
  try {
    const res = await fetch('/api/recipes/edit', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id: recipe.id, name: editNameValue.value, notes: recipe.notes || '' })
    })
    if (res.ok) {
      recipe.name = editNameValue.value
    }
  } catch (e) { console.error(e) }
  editingNameId.value = null
}

const saveNotes = async (recipe) => {
  try {
    await fetch('/api/recipes/edit', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id: recipe.id, name: recipe.name, notes: recipe.notes || '' })
    })
  } catch (e) { console.error(e) }
}

// ---- Ingredient Quantity Editing ----
const startEditQty = async (ing) => {
  editingQty.value = ing
  editQtyValue.value = ing.quantity
  await nextTick()
  const inputs = qtyEditInput.value
  if (Array.isArray(inputs) && inputs.length) inputs[0].select()
  else if (inputs) inputs.select()
}

const saveIngredientQty = async (recipeId, ing) => {
  if (!editQtyValue.value || editQtyValue.value <= 0 || editQtyValue.value === ing.quantity) {
    editingQty.value = null
    return
  }
  try {
    const res = await fetch('/api/recipes/ingredients/edit', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ recipe_id: recipeId, ingredient_id: ing.ingredient_id, quantity: editQtyValue.value })
    })
    if (res.ok) {
      ing.quantity = editQtyValue.value
      // Also update the pill tags
      const recipe = recipes.value.find(r => r.id === recipeId)
      if (recipe) {
        const tagIng = (recipe._ingredients || []).find(i => i.ingredient_id === ing.ingredient_id)
        if (tagIng) tagIng.quantity = editQtyValue.value
      }
    }
  } catch (e) { console.error(e) }
  editingQty.value = null
}

// ---- Existing functions ----
const fetchRecipes = async () => {
    try {
        const res = await fetch('/api/recipes')
        if (res.ok) {
            const data = (await res.json()) || []
            for (const r of data) {
                try {
                    const ires = await fetch(`/api/recipes/ingredients?recipe_id=${r.id}`)
                    if (ires.ok) r._ingredients = (await ires.json()) || []
                } catch (e) { r._ingredients = [] }
            }
            recipes.value = data
        }
    } catch (e) { console.error(e) }
}

const addRecipe = async () => {
    if (!newRecipeName.value) return
    try {
        const res = await fetch('/api/recipes', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name: newRecipeName.value })
        })
        if (res.ok) {
            await fetchRecipes()
            newRecipeName.value = ''
        }
    } catch (e) { console.error(e) }
}

const deleteRecipe = async (recipe) => {
    if (!confirm(`Delete "${recipe.name}"?`)) return
    try {
        const res = await fetch('/api/recipes/delete', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ id: recipe.id })
        })
        if (res.ok) {
            if (expandedId.value === recipe.id) expandedId.value = null
            await fetchRecipes()
        }
    } catch (e) { console.error(e) }
}

const toggleRecipe = async (id) => {
    if (expandedId.value === id) {
        expandedId.value = null
        return
    }
    expandedId.value = id
    await fetchRecipeIngredients(id)
}

const fetchRecipeIngredients = async (recipeId) => {
    try {
        const res = await fetch(`/api/recipes/ingredients?recipe_id=${recipeId}`)
        if (res.ok) recipeIngredients.value = (await res.json()) || []
    } catch (e) { console.error(e) }
}

const addIngredientToRecipe = async (recipeId) => {
    if (!addIngForm.value.new_name) return

    try {
        const res = await fetch('/api/recipes/ingredients', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                recipe_id: recipeId,
                ingredient_name: addIngForm.value.new_name,
                quantity: addIngForm.value.quantity,
                is_tracked: addIngForm.value.is_tracked
            })
        })
        if (res.ok) {
            await fetchRecipeIngredients(recipeId)
            const recipe = recipes.value.find(r => r.id === recipeId)
            if (recipe) recipe._ingredients = [...recipeIngredients.value]
            addIngForm.value = { new_name: '', quantity: 1, is_tracked: true }
        }
    } catch (e) { console.error(e) }
}

const removeIngredient = async (recipeId, ingredientId) => {
    try {
        const res = await fetch('/api/recipes/ingredients/remove', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ recipe_id: recipeId, ingredient_id: ingredientId })
        })
        if (res.ok) {
            await fetchRecipeIngredients(recipeId)
            const recipe = recipes.value.find(r => r.id === recipeId)
            if (recipe) recipe._ingredients = [...recipeIngredients.value]
        }
    } catch (e) { console.error(e) }
}

onMounted(fetchRecipes)
</script>
