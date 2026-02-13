<template>
  <div class="p-6">
    <div class="flex justify-between items-center mb-6">
      <h2 class="text-2xl font-bold text-gray-800">🛒 Shopping List</h2>
      <button @click="sendToLine" :disabled="sending || shoppingList.length === 0"
              class="bg-[#06C755] text-white px-4 py-2 rounded-lg font-bold shadow-md hover:bg-[#05b34c] transition-colors flex items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed">
        <span>{{ sending ? 'Sending...' : 'Send to LINE' }}</span>
        <span v-if="!sending">💬</span>
      </button>
    </div>

    <div v-if="loading" class="text-gray-400 text-center py-10">Loading...</div>
    
    <div v-else-if="shoppingList.length === 0" class="text-center py-16">
      <div class="text-6xl mb-4">✅</div>
      <p class="text-green-500 font-semibold text-lg">Everything is in stock!</p>
      <p class="text-gray-400 text-sm mt-1">No shopping needed for your meal plan.</p>
    </div>

    <div v-else>
      <div class="bg-white rounded-2xl shadow-lg border border-gray-100 overflow-hidden">
        <table class="w-full text-left">
          <thead class="bg-gray-50">
            <tr>
              <th class="p-4 text-xs font-semibold text-gray-400 uppercase tracking-wider">Item</th>
              <th class="p-4 text-xs font-semibold text-gray-400 uppercase tracking-wider">Need to Buy</th>
              <th class="p-4 text-xs font-semibold text-gray-400 uppercase tracking-wider">Est. Cost</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(item, index) in shoppingList" :key="index" class="border-t border-gray-50 hover:bg-gray-50 transition-colors">
              <td class="p-4 font-medium text-gray-800">{{ item.name }}</td>
              <td class="p-4">
                <span class="bg-red-100 text-red-600 px-2 py-0.5 rounded-full text-sm font-semibold">{{ item.quantity_needed }}</span>
              </td>
              <td class="p-4 text-gray-600">${{ Math.round(item.estimated_cost) }}</td>
            </tr>
          </tbody>
          <tfoot>
            <tr class="bg-gradient-to-r from-gray-50 to-blue-50 border-t-2 border-gray-200">
              <td class="p-4 font-bold text-gray-700 text-right">Total:</td>
              <td class="p-4"></td>
              <td class="p-4 font-bold text-lg text-blue-600">${{ Math.round(totalCost) }}</td>
            </tr>
          </tfoot>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'

const shoppingList = ref([])
const loading = ref(true)
const sending = ref(false)

const sendToLine = async () => {
  if (shoppingList.value.length === 0) return
  sending.value = true
  try {
    const res = await fetch('http://localhost:8080/api/line/send-shopping-list', { method: 'POST' })
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

const fetchShoppingList = async () => {
  try {
    const res = await fetch('http://localhost:8080/api/shopping-list')
    if (res.ok) {
      const data = await res.json()
      shoppingList.value = data || []
    }
  } catch (e) {
    console.error("Failed to fetch shopping list", e)
  } finally {
    loading.value = false
  }
}

const totalCost = computed(() => {
  return shoppingList.value.reduce((sum, item) => sum + item.estimated_cost, 0)
})

onMounted(fetchShoppingList)
</script>
