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
    <div class="grid grid-cols-7 gap-3">
      <!-- Day Headers -->
      <div v-for="day in weekDays" :key="day.dateStr"
           class="text-center">
        <div class="text-xs uppercase tracking-wider font-semibold text-gray-400 mb-1">{{ day.dayName }}</div>
        <div :class="[day.isToday ? 'bg-blue-500 text-white' : 'bg-gray-100 text-gray-600']"
             class="rounded-full w-8 h-8 flex items-center justify-center mx-auto text-sm font-bold mb-2">
          {{ day.dayNum }}
        </div>

        <!-- Day Card -->
        <div class="bg-white rounded-xl border border-gray-100 shadow-sm min-h-[140px] p-2 space-y-1">
          <!-- Meals for this day -->
          <div v-for="meal in getMealsForDay(day.dateStr)" :key="meal.id"
               class="group relative rounded-lg p-2 text-left transition-all"
               :class="meal.meal_type === 'Lunch' ? 'bg-amber-50 border border-amber-100' : 'bg-indigo-50 border border-indigo-100'">
            <div class="text-[10px] uppercase tracking-wider font-bold"
                 :class="meal.meal_type === 'Lunch' ? 'text-amber-500' : 'text-indigo-500'">
              {{ meal.meal_type === 'Lunch' ? '☀️' : '🌙' }} {{ meal.meal_type }}
            </div>
            <div class="text-xs font-semibold text-gray-700 truncate">{{ meal.recipe_name }}</div>
            <!-- Delete button -->
            <button @click="deleteMeal(meal.id)"
                    class="absolute top-1 right-1 opacity-0 group-hover:opacity-100 text-red-400 hover:text-red-600 transition-all text-xs leading-none">✕</button>
          </div>

          <!-- Add Meal Button -->
          <button @click="openAdd(day.dateStr)"
                  class="w-full text-center text-gray-300 hover:text-blue-500 hover:bg-blue-50 rounded-lg py-1 transition-all text-lg">+</button>
        </div>
      </div>
    </div>

    <!-- Add/Edit Modal -->
    <div v-if="showModal" class="fixed inset-0 bg-black/30 backdrop-blur-sm flex items-center justify-center z-50" @click.self="showModal = false">
      <div class="bg-white rounded-2xl shadow-2xl p-6 w-80">
        <h3 class="font-bold text-lg mb-4 text-gray-800">Schedule Meal</h3>
        <p class="text-sm text-gray-400 mb-3">{{ formatDateDisplay(modalDate) }}</p>

        <div class="mb-3">
          <label class="block text-xs font-semibold text-gray-500 uppercase tracking-wider mb-1">Meal Type</label>
          <div class="flex gap-2">
            <button @click="modalMealType = 'Lunch'"
                    :class="modalMealType === 'Lunch' ? 'bg-amber-500 text-white' : 'bg-gray-100 text-gray-600'"
                    class="flex-1 py-2 rounded-lg font-medium text-sm transition-colors">☀️ Lunch</button>
            <button @click="modalMealType = 'Dinner'"
                    :class="modalMealType === 'Dinner' ? 'bg-indigo-500 text-white' : 'bg-gray-100 text-gray-600'"
                    class="flex-1 py-2 rounded-lg font-medium text-sm transition-colors">🌙 Dinner</button>
          </div>
        </div>

        <div class="mb-4">
          <label class="block text-xs font-semibold text-gray-500 uppercase tracking-wider mb-1">Recipe</label>
          <select v-model="modalRecipeId" class="border-2 border-gray-200 p-2 rounded-lg w-full text-sm focus:ring-2 focus:ring-blue-400 focus:border-blue-400 focus:outline-none">
            <option value="" disabled>Select recipe...</option>
            <option v-for="r in recipes" :key="r.id" :value="r.id">{{ r.name }}</option>
          </select>
        </div>

        <div class="flex gap-2">
          <button @click="showModal = false"
                  class="flex-1 py-2 rounded-lg bg-gray-100 text-gray-600 hover:bg-gray-200 transition-colors font-medium text-sm">Cancel</button>
          <button @click="saveMeal"
                  class="flex-1 py-2 rounded-lg bg-gradient-to-r from-blue-500 to-indigo-500 text-white hover:from-blue-600 hover:to-indigo-600 transition-all font-medium text-sm shadow-md">Save</button>
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
const modalMealType = ref('Dinner')
const modalRecipeId = ref('')

const getMonday = (offset) => {
  const d = new Date()
  const day = d.getDay()
  const diff = d.getDate() - day + (day === 0 ? -6 : 1) + (offset * 7)
  const monday = new Date(d.setDate(diff))
  monday.setHours(0, 0, 0, 0)
  return monday
}

const weekDays = computed(() => {
  const monday = getMonday(weekOffset.value)
  const days = []
  const dayNames = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']
  const today = new Date().toISOString().split('T')[0]

  for (let i = 0; i < 7; i++) {
    const d = new Date(monday)
    d.setDate(monday.getDate() + i)
    const dateStr = d.toISOString().split('T')[0]
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
      if (a.meal_type === 'Lunch' && b.meal_type === 'Dinner') return -1
      if (a.meal_type === 'Dinner' && b.meal_type === 'Lunch') return 1
      return 0
    })
}

const fetchRecipes = async () => {
  try {
    const res = await fetch('http://localhost:8080/api/recipes')
    if (res.ok) recipes.value = (await res.json()) || []
  } catch (e) { console.error(e) }
}

const fetchMealPlan = async () => {
  try {
    const res = await fetch('http://localhost:8080/api/mealplan')
    if (res.ok) mealPlan.value = (await res.json()) || []
  } catch (e) { console.error(e) }
}

const openAdd = (dateStr) => {
  modalDate.value = dateStr
  modalMealType.value = 'Dinner'
  modalRecipeId.value = ''
  showModal.value = true
}

const saveMeal = async () => {
  if (!modalRecipeId.value) return
  try {
    const res = await fetch('http://localhost:8080/api/mealplan', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        date: modalDate.value,
        meal_type: modalMealType.value,
        recipe_id: parseInt(modalRecipeId.value)
      })
    })
    if (res.ok) {
      await fetchMealPlan()
      showModal.value = false
    }
  } catch (e) { console.error(e) }
}

const deleteMeal = async (id) => {
  try {
    const res = await fetch('http://localhost:8080/api/mealplan/delete', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id })
    })
    if (res.ok) await fetchMealPlan()
  } catch (e) { console.error(e) }
}

const prevWeek = () => weekOffset.value--
const nextWeek = () => weekOffset.value++

const formatDateDisplay = (dateStr) => {
  return new Date(dateStr).toLocaleDateString('en-US', { weekday: 'long', month: 'long', day: 'numeric' })
}

onMounted(() => {
  fetchRecipes()
  fetchMealPlan()
})
</script>
