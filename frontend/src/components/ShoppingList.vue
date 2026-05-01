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

    <div v-if="loading" class="space-y-3 mt-2">
      <div v-for="i in 3" :key="i" class="h-14 bg-gray-100 rounded-2xl animate-pulse" />
    </div>

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
      <p class="text-gray-400 text-sm mt-1">All tracked items are in stock.</p>
    </div>

    <div v-else-if="!loading">
      <div class="space-y-2">
        <!-- Auto-generated low stock items -->
        <div v-for="(item, index) in displayList" :key="item.id || 'item-'+index"
             :class="[item.is_checked ? 'opacity-50' : '', item.is_custom ? 'border-blue-100 bg-blue-50/30' : 'border-gray-100 bg-white']"
             class="flex items-center gap-3 p-4 rounded-xl border shadow-sm transition-all hover:shadow-md group">
          <input type="checkbox" v-model="item.is_checked" @change="saveList(item)"
                 class="w-5 h-5 rounded-lg text-blue-500 focus:ring-blue-400 cursor-pointer flex-shrink-0" />
          <div class="flex items-center gap-1.5 flex-1 min-w-0">
            <p :class="item.is_checked ? 'line-through text-gray-400' : 'text-gray-800'" class="font-medium text-sm truncate">
              {{ item.name }}
            </p>
            <AppBadge :variant="item.is_custom ? 'custom' : 'auto'" class="flex-shrink-0">
              {{ item.is_custom ? 'Custom' : 'Auto' }}
            </AppBadge>
          </div>
          <button @click="removeItem(index)"
                  class="text-gray-300 hover:text-red-500 active:text-red-600 transition-colors text-xl leading-none flex-shrink-0 p-1">×</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useToast } from '../composables/useToast.js'
import AppBadge from './ui/AppBadge.vue'

const { toast } = useToast()

const autoItems = ref([])
const customItems = ref([])
const loading = ref(true)
const sending = ref(false)
const newItem = ref('')

// Merge auto + custom items for display
const displayList = computed(() => {
  return [...autoItems.value, ...customItems.value]
})



const addCustomItem = async () => {
  const name = newItem.value.trim()
  if (!name) return
  
  try {
    const res = await fetch('/api/shopping-list', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name })
    })
    if (res.ok) {
      newItem.value = ''
      await fetchShoppingList()
    }
  } catch (e) {
    toast('Failed to add item', 'error')
  }
}

const removeItem = async (index) => {
  const item = displayList.value[index]
  if (!item || !item.id) return

  try {
    const res = await fetch('/api/shopping-list/delete', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id: item.id })
    })
    if (res.ok) {
      await fetchShoppingList()
    }
  } catch (e) {
    toast('Failed to delete item', 'error')
  }
}

// Persist checked state to backend
const saveList = async (item) => {
  if (!item || !item.id) return

  try {
    await fetch('/api/shopping-list/check', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id: item.id, is_checked: item.is_checked })
    })
  } catch (e) {
    toast('Failed to update item', 'error')
  }
}

const fetchShoppingList = async () => {
  try {
    const res = await fetch('/api/shopping-list')
    if (res.ok) {
      const data = await res.json()
      
      autoItems.value = (data || []).filter(i => !i.is_custom)
      customItems.value = (data || []).filter(i => i.is_custom)
    }
  } catch (e) {
    toast('Failed to load shopping list', 'error')
  } finally {
    loading.value = false
  }
}

const sendToLine = async () => {
  const itemsToSend = displayList.value
    .filter(i => !i.is_checked)
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
      toast('Shopping list sent to LINE! 💬')
    } else {
      const text = await res.text()
      try {
        const err = JSON.parse(text)
        toast('Failed to send: ' + (err.message || text), 'error')
      } catch {
        toast('Failed to send: ' + text, 'error')
      }
    }
  } catch (e) {
    toast('Error sending to LINE: ' + e.message, 'error')
  } finally {
    sending.value = false
  }
}

onMounted(fetchShoppingList)
</script>
