<template>
  <div class="p-6">
    <h2 class="text-2xl font-bold mb-2 text-gray-800">📅 Meal Planner</h2>

    <!-- Week Navigation -->
    <div class="flex justify-between items-center mb-5">
      <button @click="prevWeek" class="bg-white border border-gray-200 px-4 py-2 rounded-xl hover:bg-gray-50 transition-colors shadow-sm font-medium text-gray-600">← Prev</button>
      <div class="flex flex-col items-center gap-1">
        <span class="font-semibold text-gray-600 text-sm">{{ weekLabel }}</span>
        <button
          @click="weekOffset = 0"
          :class="[
            'px-3 py-1 rounded-lg text-xs font-medium transition-colors',
            weekOffset === 0
              ? 'bg-primary-100 text-primary-700 cursor-default'
              : 'bg-gray-100 hover:bg-gray-200 text-gray-600'
          ]"
          :disabled="weekOffset === 0"
        >Today</button>
      </div>
      <button @click="nextWeek" class="bg-white border border-gray-200 px-4 py-2 rounded-xl hover:bg-gray-50 transition-colors shadow-sm font-medium text-gray-600">Next →</button>
    </div>

    <!-- Loading skeleton -->
    <div v-if="loading" class="grid grid-cols-2 md:grid-cols-7 gap-3 mt-4">
      <div v-for="i in 7" :key="i" class="h-32 bg-gray-100 rounded-2xl animate-pulse" />
    </div>

    <!-- Weekly Calendar Grid -->
    <div v-else>
      <!-- Mobile day navigation -->
      <div class="flex items-center justify-between mb-3 md:hidden">
        <button
          @click="mobileDayStart = Math.max(0, mobileDayStart - 1)"
          :disabled="mobileDayStart === 0"
          class="w-9 h-9 rounded-xl bg-gray-100 hover:bg-gray-200 disabled:opacity-30 flex items-center justify-center transition-colors"
        >←</button>
        <span class="text-sm text-gray-500 font-medium">
          {{ weekDays[mobileDayStart]?.dayName }} – {{ weekDays[Math.min(mobileDayStart + 1, 6)]?.dayName }}
        </span>
        <button
          @click="mobileDayStart = Math.min(5, mobileDayStart + 1)"
          :disabled="mobileDayStart >= 5"
          class="w-9 h-9 rounded-xl bg-gray-100 hover:bg-gray-200 disabled:opacity-30 flex items-center justify-center transition-colors"
        >→</button>
      </div>

      <!-- Day cards -->
      <div class="grid grid-cols-2 md:grid-cols-7 gap-3 pb-24 md:pb-0">
      <div v-for="(day, idx) in weekDays" :key="day.dateStr"
           :class="['text-center md:block', idx >= mobileDayStart && idx < mobileDayStart + 2 ? 'block' : 'hidden']">
        <div class="text-xs uppercase tracking-wider font-semibold text-gray-400 mb-1">{{ day.dayName }}</div>
        <div class="flex items-center justify-center gap-1 mb-2">
          <div :class="[day.isToday ? 'bg-blue-500 text-white' : 'bg-gray-100 text-gray-600']"
               class="rounded-full w-8 h-8 flex items-center justify-center text-sm font-bold">
            {{ day.dayNum }}
          </div>
          <span
            v-if="getMealsForDay(day.dateStr).length > 0"
            class="inline-flex items-center px-1.5 py-0.5 rounded-full text-[10px] font-semibold bg-primary-100 text-primary-700"
          >{{ getMealsForDay(day.dateStr).length }}</span>
        </div>

        <!-- Day Card -->
        <div class="bg-white rounded-xl border border-gray-100 shadow-sm min-h-[140px] p-3 flex flex-col gap-1.5">
          <!-- Iterate over meal types -->
          <div v-for="type in ['Breakfast', 'Lunch', 'Dinner']" :key="type">
            <!-- Show block only if there are meals for this type -->
            <div v-if="getMealsForDay(day.dateStr).filter(m => m.meal_type === type).length > 0"
                 class="rounded-lg px-2 py-1.5 transition-all border mb-1"
                 :class="mealTypeStyle(type).bg">
              <!-- Header -->
              <div class="text-[10px] uppercase tracking-wider font-bold mb-1 flex items-center gap-1"
                   :class="mealTypeStyle(type).text">
                <span>{{ mealTypeStyle(type).icon }}</span>
                <span class="hidden md:inline">{{ type }}</span>
              </div>
              <!-- Meals List -->
              <div class="space-y-1">
                <div v-for="meal in getMealsForDay(day.dateStr).filter(m => m.meal_type === type)" :key="meal.id"
                     class="group relative"
                     :title="meal.custom_name || meal.recipe_name">

                  <!-- Meal Name -->
                  <div class="text-[11px] font-medium leading-snug flex items-start gap-1"
                       :class="meal.is_cooked ? 'text-gray-400 font-normal line-through opacity-70' : 'text-gray-700'">
                    <span v-if="meal.is_cooked" class="select-none shrink-0" title="Yummy!">😋</span>
                    <span class="line-clamp-2">{{ meal.custom_name || meal.recipe_name || 'Unnamed' }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

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
        </div>
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
                   class="flex items-center justify-between bg-gray-50 p-2 rounded-lg group">
                <div class="flex items-center gap-2 flex-1 min-w-0">
                  <span class="text-sm font-medium text-gray-700 truncate">{{ meal.custom_name || getRecipeName(meal.recipe_id) }}</span>
                </div>
                <button v-if="meal.recipe_id" @click="editNewMeal(meal)" class="text-gray-300 hover:text-blue-400 p-1 transition-colors text-sm leading-none" title="Edit ingredients">✏️</button>
                <button @click="removeNewMeal(meal)" class="text-gray-300 hover:text-red-400 p-1 ml-1 transition-colors text-lg leading-none">🗑️</button>
              </div>
            </div>

            <!-- Add New Meal Controls -->
            <div class="space-y-2">
              <div v-if="!showIngredientEditor[type]" class="flex gap-2">
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

                <!-- Cancel / Add -->
                <div class="flex gap-2 pt-1">
                  <button @click="cancelIngredientEditor(type)"
                          class="flex-1 py-1.5 rounded-lg bg-white border border-gray-200 text-gray-500 hover:bg-gray-50 text-sm font-medium">Cancel</button>
                  <button @click="addNewMeal(type)"
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

    <!-- Meal Edit Sub-Modal -->
    <div v-if="showMealEditModal"
         class="fixed inset-0 bg-black/30 backdrop-blur-sm flex items-center justify-center z-[60] p-4"
         @click.self="closeMealEditModal">
      <div class="bg-white rounded-2xl shadow-2xl w-full max-w-sm flex flex-col max-h-[90vh]">
        <!-- Header -->
        <div class="p-5 border-b border-gray-100 flex items-center justify-between flex-shrink-0">
          <div>
            <h3 class="font-bold text-base text-gray-800">Edit Meal</h3>
            <p class="text-xs text-gray-400">{{ mealEditType }}</p>
          </div>
          <button @click="closeMealEditModal"
                  class="w-8 h-8 flex items-center justify-center rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition-colors"
                  aria-label="Close">
            <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
              <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd"/>
            </svg>
          </button>
        </div>
        <!-- Body -->
        <div class="p-5 overflow-y-auto flex-1 space-y-4">
          <!-- Variant Name -->
          <div>
            <label class="text-xs font-medium text-gray-500 mb-1 block">Variant name (optional)</label>
            <input v-model="variantNames[mealEditType]" type="text"
                   :placeholder="getRecipeName(selectedRecipes[mealEditType]) + ' variant...'"
                   class="w-full border border-gray-200 rounded-lg text-sm p-2 focus:outline-none focus:ring-2 focus:ring-blue-100" />
          </div>
          <!-- Ingredient List -->
          <div>
            <label class="text-xs font-medium text-gray-500 mb-1 block">Ingredients</label>
            <div class="space-y-1.5">
              <div v-for="(ing, idx) in editingIngredients[mealEditType]" :key="ing.ingredient_id"
                   class="flex items-center gap-2 bg-white rounded-md p-1.5 border border-gray-100">
                <span class="text-sm text-gray-700 flex-1 truncate">{{ ing.name }}</span>
                <input v-model.number="ing.quantity" type="number" min="1"
                       class="w-16 border border-gray-200 rounded text-sm p-1 text-center focus:outline-none focus:ring-2 focus:ring-blue-100" />
                <button @click="editingIngredients[mealEditType].splice(idx, 1)"
                        class="text-gray-300 hover:text-red-400 text-sm px-1">✕</button>
              </div>
            </div>
          </div>
          <!-- Add Ingredient -->
          <div>
            <select v-model="addIngredientSelection[mealEditType]"
                    @change="addIngredientToEditor(mealEditType)"
                    class="w-full border border-gray-200 rounded-lg text-sm p-1.5 focus:outline-none focus:ring-2 focus:ring-blue-100">
              <option value="" disabled>Add ingredient...</option>
              <option v-for="inv in availableIngredients(mealEditType)" :key="inv.id" :value="inv.id">{{ inv.name }}</option>
            </select>
          </div>
        </div>
        <!-- Footer -->
        <div class="p-4 border-t border-gray-100 bg-gray-50 rounded-b-2xl flex gap-3 flex-shrink-0">
          <button @click="closeMealEditModal"
                  class="flex-1 py-2.5 rounded-xl bg-white border border-gray-200 text-gray-600 hover:bg-gray-50 transition-colors font-semibold text-sm">Cancel</button>
          <button @click="saveEditingMeal"
                  class="flex-1 py-2.5 rounded-xl bg-green-500 text-white hover:bg-green-600 transition-colors font-semibold text-sm">Save Meal</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import ConfirmDialog from './ui/ConfirmDialog.vue'
import { useToast } from '../composables/useToast.js'

const emit = defineEmits(['inventory-changed'])
const { toast } = useToast()
const confirmCookMeal = ref(null)

const loading = ref(true)
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

const showMealEditModal = ref(false)
const mealEditType = ref('')
const editingNewMealKey = ref(null)

const confirmCookAllDayDate = ref(null)
const cookAllDayResult = ref(null)
const isCookingAll = ref(false)

// Escape key handler for meal edit sub-modal
const onMealModalKey = (e) => {
  if (e.key === 'Escape' && showMealEditModal.value) {
    e.stopImmediatePropagation()
    closeMealEditModal()
  }
}

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
  finally { loading.value = false }
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
  cancelIngredientEditor(type)

  const draftMeal = updatedMeals.value.find(m => m.id === meal.id)

  if (draftMeal) {
    editingIngredients.value[type] = draftMeal.ingredients.map(ing => ({
      ingredient_id: ing.ingredient_id,
      name: ing.name,
      quantity: ing.quantity,
    }))
  } else {
    let loaded = false
    try {
      const res = await fetch(`/api/mealplan/ingredients?meal_plan_id=${meal.id}`)
      if (res.ok) {
        const ingredients = (await res.json()) || []
        editingIngredients.value[type] = ingredients.map(ing => ({
          ingredient_id: ing.ingredient_id,
          name: ing.name,
          quantity: ing.quantity,
        }))
        loaded = true
      }
    } catch (e) { toast('Network error', 'error') }
    if (!loaded) return
  }

  editingMealId.value = meal.id
  selectedRecipes.value[type] = meal.recipe_id || ''
  variantNames.value[type] = draftMeal ? draftMeal.custom_name : (meal.custom_name || '')
  mealEditType.value = type
  showMealEditModal.value = true
}

const editNewMeal = (meal) => {
  const type = meal.meal_type
  editingIngredients.value[type] = meal.ingredients.map(ing => {
    const inv = inventory.value.find(i => i.id === ing.ingredient_id)
    return { ingredient_id: ing.ingredient_id, name: inv?.name || 'Unknown', quantity: ing.quantity }
  })
  variantNames.value[type] = meal.custom_name || ''
  selectedRecipes.value[type] = meal.recipe_id || ''
  editingMealId.value = null
  editingNewMealKey.value = meal._key
  mealEditType.value = type
  showMealEditModal.value = true
}

const saveEditingMeal = () => {
  const type = mealEditType.value
  if (!type) return

  const customName = variantNames.value[type].trim()
  const draftIngredients = editingIngredients.value[type]
    .filter(ing => ing.quantity > 0)
    .map(ing => ({ ingredient_id: ing.ingredient_id, name: ing.name, quantity: ing.quantity }))

  if (editingNewMealKey.value) {
    const idx = newMeals.value.findIndex(m => m._key === editingNewMealKey.value)
    if (idx > -1) {
      newMeals.value[idx] = {
        ...newMeals.value[idx],
        custom_name: customName,
        ingredients: draftIngredients.map(ing => ({ ingredient_id: ing.ingredient_id, quantity: ing.quantity }))
      }
    }
  } else {
    const mealId = editingMealId.value
    if (!mealId) return
    updatedMeals.value = updatedMeals.value.filter(m => m.id !== mealId)
    updatedMeals.value.push({ id: mealId, custom_name: customName, ingredients: draftIngredients })
  }

  closeMealEditModal()
}

const cancelIngredientEditor = (type) => {
  showIngredientEditor.value[type] = false
  selectedRecipes.value[type] = ''
  variantNames.value[type] = ''
  editingIngredients.value[type] = []
  addIngredientSelection.value[type] = ''
  editingMealId.value = null
}

const closeMealEditModal = () => {
  showMealEditModal.value = false
  cancelIngredientEditor(mealEditType.value)
  mealEditType.value = ''
  editingNewMealKey.value = null
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
      emit('inventory-changed')
      toast('Marked as cooked! 🍳')
    } else {
      toast('Failed to mark as cooked', 'error')
    }
  } catch (e) {
    toast('Failed to mark as cooked', 'error')
  }
}

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
    if (result.cooked.length > 0) emit('inventory-changed')
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

const mobileDayStart = ref(0)
const prevWeek = () => weekOffset.value--
const nextWeek = () => weekOffset.value++

watch(weekOffset, () => { mobileDayStart.value = 0 })

const formatDateDisplay = (dateStr) => {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleDateString('en-US', { weekday: 'long', month: 'long', day: 'numeric' })
}

onMounted(() => {
  fetchRecipes()
  fetchMealPlan()
  fetchInventory()
  document.addEventListener('keydown', onMealModalKey)
})

onUnmounted(() => {
  document.removeEventListener('keydown', onMealModalKey)
})
</script>
