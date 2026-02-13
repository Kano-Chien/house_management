<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-4">
      <h2 class="text-2xl font-bold">Inventory</h2>
      <button @click="showAddForm = !showAddForm"
        :class="[
          'flex items-center gap-1.5 px-4 py-2 rounded-lg text-sm font-medium transition-all duration-200',
          showAddForm
            ? 'bg-gray-200 text-gray-600 hover:bg-gray-300'
            : 'bg-blue-500 text-white hover:bg-blue-600 shadow-sm hover:shadow'
        ]">
        <span class="text-lg leading-none transition-transform duration-200"
          :style="{ transform: showAddForm ? 'rotate(45deg)' : 'rotate(0)' }">+</span>
        {{ showAddForm ? 'Close' : 'Add Item' }}
      </button>
    </div>

    <!-- Category Filter Toggle -->
    <div class="mb-4 flex gap-1 bg-gray-100 p-1 rounded-xl w-fit">
      <button
        v-for="opt in filterOptions" :key="opt.value"
        @click="selectedCategory = opt.value"
        :class="[
          'px-4 py-1.5 rounded-lg text-sm font-medium transition-all duration-200',
          selectedCategory === opt.value
            ? 'bg-white text-blue-600 shadow-sm'
            : 'text-gray-500 hover:text-gray-700'
        ]"
      >{{ opt.label }}</button>
    </div>

    <!-- Add Form (Collapsible) -->
    <Transition name="slide">
      <div v-if="showAddForm" class="mb-6 p-4 bg-gradient-to-r from-blue-50 to-indigo-50 rounded-xl border border-blue-100 shadow-sm">
        <div class="flex gap-2 flex-wrap items-end">
          <div class="flex flex-col gap-1">
            <label class="text-xs font-medium text-gray-500 uppercase tracking-wide">Name</label>
            <input ref="nameInputRef" v-model="newItem.name" placeholder="e.g. Chicken, Soap..."
              @keyup.enter="addItem"
              class="border border-gray-200 p-2 rounded-lg focus:ring-2 focus:ring-blue-400 focus:border-blue-400 focus:outline-none transition-all w-40" />
          </div>
          <div class="flex flex-col gap-1">
            <label class="text-xs font-medium text-gray-500 uppercase tracking-wide">Stock</label>
            <input v-model.number="newItem.current_stock" type="number" placeholder="0"
              @keyup.enter="addItem"
              class="border border-gray-200 p-2 rounded-lg focus:ring-2 focus:ring-blue-400 focus:border-blue-400 focus:outline-none transition-all w-20" />
          </div>
          <div class="flex flex-col gap-1">
            <label class="text-xs font-medium text-gray-500 uppercase tracking-wide">Price</label>
            <input v-model.number="newItem.price" type="number" step="1" min="0" placeholder="$0"
              @keyup.enter="addItem"
              class="border border-gray-200 p-2 rounded-lg focus:ring-2 focus:ring-blue-400 focus:border-blue-400 focus:outline-none transition-all w-20" />
          </div>
          <div class="flex flex-col gap-1">
            <label class="text-xs font-medium text-gray-500 uppercase tracking-wide">Category</label>
            <select v-model="newItem.category" class="border border-gray-200 p-2 rounded-lg bg-white text-sm focus:ring-2 focus:ring-blue-400 focus:outline-none transition-all">
              <option value="food">🍖 Food</option>
              <option value="daily">🧴 Daily</option>
            </select>
          </div>
          <button @click="addItem"
            :disabled="!newItem.name || adding"
            :class="[
              'px-5 py-2 rounded-lg text-sm font-medium transition-all duration-200',
              adding
                ? 'bg-green-500 text-white scale-95'
                : !newItem.name
                  ? 'bg-gray-200 text-gray-400 cursor-not-allowed'
                  : 'bg-blue-500 text-white hover:bg-blue-600 shadow-sm hover:shadow'
            ]">
            {{ adding ? '✓ Added!' : 'Add' }}
          </button>
        </div>
      </div>
    </Transition>

    <!-- Inventory Table -->
    <div class="bg-white rounded-xl shadow overflow-hidden">
      <table class="w-full text-left">
        <thead class="bg-gray-50">
            <tr>
            <th class="p-3 text-xs font-semibold text-gray-500 uppercase tracking-wide">Name</th>
            <th class="p-3 text-xs font-semibold text-gray-500 uppercase tracking-wide">Category</th>
            <th class="p-3 text-xs font-semibold text-gray-500 uppercase tracking-wide">Stock</th>
            <th class="p-3 text-xs font-semibold text-gray-500 uppercase tracking-wide">Consume</th>
            <th class="p-3 text-xs font-semibold text-gray-500 uppercase tracking-wide">Price</th>
            <th class="p-3 text-xs font-semibold text-gray-500 uppercase tracking-wide">Actions</th>
            </tr>
        </thead>
        <tbody>
            <tr v-for="item in filteredInventory" :key="item.id"
              :class="['border-t transition-colors duration-150', justAdded === item.id ? 'bg-green-50' : 'hover:bg-gray-50']">
            <!-- Editing Mode -->
            <template v-if="editingId === item.id">
                <td class="p-3"><input v-model="editForm.name" class="border p-1 rounded-lg w-full focus:ring-2 focus:ring-blue-400 focus:outline-none" /></td>
                <td class="p-3">
                  <select v-model="editForm.category" class="border p-1 rounded-lg bg-white text-sm focus:ring-2 focus:ring-blue-400 focus:outline-none">
                    <option value="food">🍖 Food</option>
                    <option value="daily">🧴 Daily</option>
                  </select>
                </td>
                <td class="p-3"><input v-model.number="editForm.current_stock" type="number" class="border p-1 rounded-lg w-20 focus:ring-2 focus:ring-blue-400 focus:outline-none" /></td>
                <td class="p-3 text-gray-400">{{ item.planned_consumption }}</td>
                <td class="p-3"><input v-model.number="editForm.price" type="number" step="1" min="0" class="border p-1 rounded-lg w-20 focus:ring-2 focus:ring-blue-400 focus:outline-none" /></td>
                <td class="p-3">
                    <div class="flex gap-1">
                    <button @click="saveEdit(item)" class="bg-green-500 text-white px-3 py-1 rounded-lg text-sm hover:bg-green-600 transition-colors">Save</button>
                    <button @click="cancelEdit" class="bg-gray-200 text-gray-600 px-3 py-1 rounded-lg text-sm hover:bg-gray-300 transition-colors">Cancel</button>
                    </div>
                </td>
            </template>
            <!-- Display Mode -->
            <template v-else>
                <td class="p-3 font-medium">{{ item.name }}</td>
                <td class="p-3">
                  <span :class="[
                    'inline-block px-2 py-0.5 rounded-full text-xs font-medium',
                    item.category === 'food'
                      ? 'bg-orange-100 text-orange-700'
                      : 'bg-blue-100 text-blue-700'
                  ]">{{ item.category === 'food' ? '🍖 Food' : '🧴 Daily' }}</span>
                </td>
                <td class="p-3">
                    <div class="flex items-center gap-2">
                    <button @click="updateStock(item, -1)" class="bg-gray-200 w-6 h-6 flex items-center justify-center rounded hover:bg-gray-300 transition-colors text-sm">-</button>
                    <span :class="{'text-red-500 font-bold': item.current_stock < item.planned_consumption}" class="min-w-[2rem] text-center">{{ item.current_stock }}</span>
                    <button @click="updateStock(item, 1)" class="bg-gray-200 w-6 h-6 flex items-center justify-center rounded hover:bg-gray-300 transition-colors text-sm">+</button>
                    </div>
                </td>
                <td class="p-3 text-gray-500">{{ item.planned_consumption }}</td>
                <td class="p-3">${{ item.price }}</td>
                <td class="p-3">
                    <div class="flex gap-1">
                    <button @click="startEdit(item)" class="bg-yellow-400 text-white px-2 py-1 rounded-lg text-sm hover:bg-yellow-500 transition-colors">Edit</button>
                    <button @click="deleteItem(item)" class="bg-red-500 text-white px-2 py-1 rounded-lg text-sm hover:bg-red-600 transition-colors">Delete</button>
                    </div>
                </td>
            </template>
            </tr>
            <tr v-if="filteredInventory.length === 0">
              <td colspan="6" class="p-8 text-center text-gray-400">
                <div class="text-3xl mb-2">📦</div>
                No items in this category
              </td>
            </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick, watch } from 'vue'

const inventory = ref([])
const newItem = ref({ name: '', current_stock: 0, price: 0, category: 'food' })
const editingId = ref(null)
const editForm = ref({ name: '', current_stock: 0, price: 0, category: 'food' })
const showAddForm = ref(false)
const adding = ref(false)
const justAdded = ref(null)
const nameInputRef = ref(null)

// Auto-focus name input when form opens
watch(showAddForm, async (val) => {
  if (val) {
    await nextTick()
    nameInputRef.value?.focus()
  }
})

// Category filter
const selectedCategory = ref('all')
const filterOptions = [
  { label: '📋 All', value: 'all' },
  { label: '🍖 Food', value: 'food' },
  { label: '🧴 Daily', value: 'daily' },
]
const filteredInventory = computed(() => {
  if (selectedCategory.value === 'all') return inventory.value
  return inventory.value.filter(item => item.category === selectedCategory.value)
})

const fetchInventory = async () => {
    try {
        const res = await fetch('http://localhost:8080/api/inventory')
        if (res.ok) {
            inventory.value = await res.json()
        }
    } catch (e) {
        console.error("Failed to fetch inventory", e)
    }
}

const addItem = async () => {
    if (!newItem.value.name || adding.value) return
    adding.value = true
    try {
        const res = await fetch('http://localhost:8080/api/inventory', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(newItem.value)
        })
        if (res.ok) {
            const added = await res.json()
            await fetchInventory()
            // Flash the new row green briefly
            justAdded.value = added.id
            setTimeout(() => { justAdded.value = null }, 1500)
            // Reset form but keep category selection
            newItem.value = { name: '', current_stock: 0, price: 0, category: newItem.value.category }
            await nextTick()
            nameInputRef.value?.focus()
        }
    } catch (e) {
        console.error("Failed to add item", e)
    } finally {
      setTimeout(() => { adding.value = false }, 600)
    }
}

const updateStock = async (item, change) => {
    const newStock = item.current_stock + change
    if (newStock < 0) return

    try {
        const res = await fetch('http://localhost:8080/api/inventory/stock', {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ id: item.id, new_stock: newStock })
        })
        if (res.ok) {
            item.current_stock = newStock
        }
    } catch (e) {
        console.error("Failed to update stock", e)
    }
}

const startEdit = (item) => {
    editingId.value = item.id
    editForm.value = { name: item.name, current_stock: item.current_stock, price: item.price, category: item.category || 'food' }
}

const cancelEdit = () => {
    editingId.value = null
}

const saveEdit = async (item) => {
    try {
        const res = await fetch('http://localhost:8080/api/inventory/edit', {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ id: item.id, ...editForm.value })
        })
        if (res.ok) {
            await fetchInventory()
            editingId.value = null
        }
    } catch (e) {
        console.error("Failed to edit item", e)
    }
}

const deleteItem = async (item) => {
    if (!confirm(`Delete "${item.name}"?`)) return
    try {
        const res = await fetch('http://localhost:8080/api/inventory/delete', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ id: item.id })
        })
        if (res.ok) {
            await fetchInventory()
        }
    } catch (e) {
        console.error("Failed to delete item", e)
    }
}

onMounted(fetchInventory)
</script>

<style scoped>
.slide-enter-active,
.slide-leave-active {
  transition: all 0.25s ease;
  overflow: hidden;
}
.slide-enter-from,
.slide-leave-to {
  opacity: 0;
  max-height: 0;
  padding-top: 0;
  padding-bottom: 0;
  margin-bottom: 0;
}
.slide-enter-to,
.slide-leave-from {
  opacity: 1;
  max-height: 200px;
}
</style>
