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
                  <div class="text-xs font-medium leading-tight"
                       :class="meal.is_cooked ? 'text-gray-400 font-normal line-through opacity-70' : 'text-gray-700'"
                       :title="meal.recipe_name">
                    {{ meal.recipe_name }}
                  </div>

                  <!-- Status Icon (Always visible if cooked) -->
                  <div v-if="meal.is_cooked" class="absolute top-1 right-1 text-xs opacity-50 grayscale cursor-default select-none">✅</div>

                  <!-- Hover Actions (Overlay) -->
                  <div class="absolute top-0.5 right-0.5 flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity bg-white/95 rounded px-1 py-0.5 shadow-sm border border-gray-100 z-10">
                    <!-- Cook Button -->
                    <button v-if="!meal.is_cooked" @click.stop="cookMeal(meal)"
                            class="text-sm hover:scale-110 leading-none text-amber-500 hover:text-amber-600 transition-transform" title="Mark as Cooked">🧑‍🍳</button>
                    
                    <!-- Delete Button -->
                    <button @click.stop="deleteMeal(meal.id)"
                            class="text-red-300 hover:text-red-500 text-xs font-bold px-1 leading-none transition-colors">✕</button>
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
                <span class="text-sm font-medium text-gray-700 truncate flex-1" :class="{'line-through text-gray-400': meal.is_cooked}">{{ meal.recipe_name }}</span>
                <button @click="markForRemoval(meal.id)" class="text-gray-400 hover:text-red-500 p-1">✕</button>
              </div>
              
              <!-- Newly Added Meals (Pending Save) -->
              <div v-for="(meal, idx) in newMeals.filter(m => m.meal_type === type)" :key="'new-'+idx"
                   class="flex items-center justify-between bg-blue-50 border border-blue-100 p-2 rounded-lg">
                <span class="text-sm font-medium text-blue-700 truncate flex-1">{{ getRecipeName(meal.recipe_id) }}</span>
                <button @click="removeNewMeal(meal)" class="text-blue-400 hover:text-blue-600 p-1">✕</button>
              </div>
            </div>

            <!-- Add New Meal Controls -->
            <div class="flex gap-2">
              <select v-model="selectedRecipes[type]" class="flex-1 border border-gray-200 rounded-lg text-sm p-1.5 focus:outline-none focus:ring-2 focus:ring-blue-100">
                <option value="" disabled>Add recipe...</option>
                <option v-for="r in recipes" :key="r.id" :value="r.id">{{ r.name }}</option>
              </select>
              <button @click="addNewMeal(type)" :disabled="!selectedRecipes[type]"
                      class="bg-gray-100 hover:bg-gray-200 text-gray-600 rounded-lg px-3 py-1.5 text-sm font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed">Add</button>
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
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'

const mealPlan = ref([])
const recipes = ref([])
const weekOffset = ref(0)
const showModal = ref(false)
const modalDate = ref('')
const isSaving = ref(false)

// Editor State
const removedMealIds = ref(new Set())
const newMeals = ref([]) // { meal_type, recipe_id }
const selectedRecipes = ref({ Breakfast: '', Lunch: '', Dinner: '' })

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
      const mDate = new Date(m.date).toISOString().split('T')[0]
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

const fetchRecipes = async () => {
  try {
    const res = await fetch('/api/recipes')
    if (res.ok) recipes.value = (await res.json()) || []
  } catch (e) { console.error(e) }
}

const fetchMealPlan = async () => {
  try {
    const res = await fetch('/api/mealplan')
    if (res.ok) mealPlan.value = (await res.json()) || []
  } catch (e) { console.error(e) }
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
  selectedRecipes.value = { Breakfast: '', Lunch: '', Dinner: '' }
  showModal.value = true
}

const closeModal = () => showModal.value = false

const markForRemoval = (id) => removedMealIds.value.add(id)
const addNewMeal = (type) => {
  const id = selectedRecipes.value[type]
  if (!id) return
  newMeals.value.push({ meal_type: type, recipe_id: id })
  selectedRecipes.value[type] = '' // reset dropdown
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

    // Add new meals
    for (const meal of newMeals.value) {
      promises.push(fetch('/api/mealplan', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          date: modalDate.value,
          meal_type: meal.meal_type,
          recipe_id: parseInt(meal.recipe_id)
        })
      }))
    }

    await Promise.all(promises)
    await fetchMealPlan()
    closeModal()
  } catch (e) { 
    console.error(e)
    alert('Error saving changes')
  } finally {
    isSaving.value = false
  }
}

const deleteMeal = async (id) => {
  if(!confirm('Delete this meal?')) return
  try {
    const res = await fetch('/api/mealplan/delete', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id })
    })
    if (res.ok) await fetchMealPlan()
  } catch (e) { console.error(e) }
}

const cookMeal = async (meal) => {
  if (!confirm(`Mark "${meal.recipe_name}" as cooked? This will deduct ingredients from inventory.`)) return
  try {
    const res = await fetch('/api/mealplan/cook', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id: meal.id })
    })
    if (res.ok) {
      await fetchMealPlan()
    } else {
      const txt = await res.text()
      alert('Failed: ' + txt)
    }
  } catch (e) { console.error(e) }
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
})
</script>
