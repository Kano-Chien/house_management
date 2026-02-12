<template>
  <div class="p-6">
    <h2 class="text-2xl font-bold mb-4">Shopping List</h2>

    <div v-if="loading" class="text-gray-600">Loading...</div>
    <div v-else-if="shoppingList.length === 0" class="text-green-600 font-medium">
      Everything is in stock! No shopping needed.
    </div>

    <div v-else class="bg-white rounded shadow overflow-hidden">
      <table class="w-full text-left">
        <thead class="bg-gray-100">
          <tr>
            <th class="p-3">Item</th>
            <th class="p-3">To Buy</th>
            <th class="p-3">Unit</th>
            <th class="p-3">Est. Cost</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(item, index) in shoppingList" :key="index" class="border-t">
            <td class="p-3 font-medium">{{ item.name }}</td>
            <td class="p-3">{{ item.quantity_needed }}</td>
            <td class="p-3">{{ item.unit }}</td>
            <td class="p-3">${{ item.estimated_cost.toFixed(2) }}</td>
          </tr>
          <tr class="bg-gray-50 font-bold border-t-2">
            <td colspan="3" class="p-3 text-right">Total Estimated Cost:</td>
            <td class="p-3">${{ totalCost.toFixed(2) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'

const shoppingList = ref([])
const loading = ref(true)

const fetchShoppingList = async () => {
  try {
    const res = await fetch('http://localhost:8080/api/shopping-list')
    if (res.ok) {
      const data = await res.json()
      // Handle null response if list is empty
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
