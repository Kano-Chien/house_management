<template>
  <div class="p-6">
    <div class="flex justify-between items-center mb-6">
      <h2 class="text-2xl font-bold text-gray-800">🛒 Shopping List</h2>
      <button @click="sendToLine" :disabled="sending || displayList.length === 0"
              class="bg-[#06C755] text-white px-4 py-2 rounded-lg font-bold shadow-md hover:bg-[#05b34c] transition-colors flex items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed">
        <span>{{ sending ? 'Sending...' : 'Send to LINE' }}</span>
        <span v-if="!sending">💬</span>
      </button>
    </div>

    <div v-if="loading" class="text-gray-400 text-center py-10">Loading...</div>

    <!-- Add custom item -->
    <div class="mb-4">
      <div class="flex gap-2">
        <input v-model="newItem" placeholder="Add item to shopping list..."
          @keyup.enter="addCustomItem"
          class="border-2 border-gray-200 p-3 rounded-xl text-sm flex-1 focus:ring-2 focus:ring-blue-400 focus:border-blue-400 focus:outline-none transition-all placeholder-gray-300" />
        <button @click="addCustomItem"
          class="bg-gradient-to-r from-blue-500 to-indigo-500 text-white px-5 py-3 rounded-xl hover:from-blue-600 hover:to-indigo-600 transition-all text-sm font-semibold shadow-md hover:shadow-lg active:scale-[0.98]">
          + Add
        </button>
      </div>
    </div>

    <div v-if="!loading && displayList.length === 0" class="text-center py-16">
      <div class="text-6xl mb-4">✅</div>
      <p class="text-green-500 font-semibold text-lg">Everything is in stock!</p>
      <p class="text-gray-400 text-sm mt-1">All tracked items have 3+ in stock.</p>
    </div>

    <div v-else-if="!loading">
      <div class="space-y-2">
        <!-- Auto-generated low stock items -->
        <div v-for="(item, index) in displayList" :key="'item-'+index"
             :class="[!item.checked ? 'opacity-40' : '', item.isCustom ? 'border-blue-100 bg-blue-50/30' : 'border-gray-100 bg-white']"
             class="flex items-center gap-3 p-4 rounded-xl border shadow-sm transition-all hover:shadow-md group">
          <input type="checkbox" v-model="item.checked" @change="saveList"
                 class="w-5 h-5 rounded-lg text-blue-500 focus:ring-blue-400 cursor-pointer flex-shrink-0" />
          <div class="flex-1 min-w-0">
            <p :class="!item.checked ? 'line-through text-gray-400' : 'text-gray-800'" class="font-medium text-sm truncate">
              {{ item.name }}
            </p>
          </div>
          <button @click="removeItem(index)"
                  class="opacity-0 group-hover:opacity-100 text-red-400 hover:text-red-600 transition-all text-xl leading-none flex-shrink-0">×</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'

const autoItems = ref([])
const customItems = ref([])
const loading = ref(true)
const sending = ref(false)
const newItem = ref('')

// Merge auto + custom items for display
const displayList = computed(() => {
  return [...autoItems.value, ...customItems.value]
})



const addCustomItem = () => {
  const name = newItem.value.trim()
  if (!name) return
  customItems.value.push({ name, checked: true, isCustom: true, estimated_cost: 0 })
  newItem.value = ''
  saveList()
}

const removeItem = (index) => {
  const list = displayList.value
  const item = list[index]
  if (item.isCustom) {
    const ci = customItems.value.indexOf(item)
    if (ci >= 0) customItems.value.splice(ci, 1)
  } else {
    const ai = autoItems.value.indexOf(item)
    if (ai >= 0) autoItems.value.splice(ai, 1)
  }
  saveList()
}

// Persist custom items and checked state to localStorage
const saveList = () => {
  const state = {
    custom: customItems.value,
    checked: {}
  }
  autoItems.value.forEach(item => {
    if (item.checked) state.checked[item.name] = true
  })
  localStorage.setItem('shopping_list_state', JSON.stringify(state))
}

const loadSavedState = () => {
  try {
    const raw = localStorage.getItem('shopping_list_state')
    if (!raw) return
    const state = JSON.parse(raw)
    if (state.custom) customItems.value = state.custom
    if (state.checked) {
      autoItems.value.forEach(item => {
        if (state.checked[item.name]) item.checked = true
      })
    }
  } catch (e) { console.error(e) }
}

const fetchShoppingList = async () => {
  try {
    const res = await fetch('/api/shopping-list')
    if (res.ok) {
      const data = await res.json()
      autoItems.value = (data || []).map(item => ({

        ...item,
        checked: true,
        isCustom: false,
      }))
      loadSavedState()
    }
  } catch (e) {
    console.error("Failed to fetch shopping list", e)
  } finally {
    loading.value = false
  }
}

const sendToLine = async () => {
  // checked = user wants to buy
  const itemsToSend = displayList.value
    .filter(i => i.checked)
    .map(i => ({ name: i.name }))
  
  if (itemsToSend.length === 0) return
  sending.value = true
  try {
    const res = await fetch('/api/line/send-shopping-list', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(itemsToSend)
    })
    if (res.ok) {
      alert('Shopping list broadcasted to LINE!')
    } else {
      const text = await res.text()
      try {
        const err = JSON.parse(text)
        alert('Failed: ' + (err.message || text))
      } catch (e) {
        alert('Failed: ' + text)
      }
    }
  } catch (e) {
    console.error(e)
    alert('Error sending to LINE: ' + e.message)
  } finally {
    sending.value = false
  }
}

onMounted(fetchShoppingList)
</script>
