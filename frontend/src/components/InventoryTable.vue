<template>
  <div class="p-6">
    <h2 class="text-2xl font-bold mb-4">Inventory</h2>

    <!-- Add Form -->
    <div class="mb-6 p-4 bg-white rounded shadow">
      <h3 class="font-semibold mb-2">Add Ingredient</h3>
      <div class="flex gap-2">
        <input v-model="newItem.name" placeholder="Name" class="border p-2 rounded" />
        <input v-model.number="newItem.current_stock" type="number" placeholder="Stock" class="border p-2 rounded w-24" />
        <input v-model.number="newItem.price" type="number" step="0.01" placeholder="Price" class="border p-2 rounded w-24" />
        <button @click="addItem" class="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600">Add</button>
      </div>
    </div>

    <!-- Inventory Table -->
    <div class="bg-white rounded shadow overflow-hidden">
      <table class="w-full text-left">
        <thead class="bg-gray-100">
            <tr>
            <th class="p-3">Name</th>
            <th class="p-3">Stock</th>
            <th class="p-3">Consume</th>
            <th class="p-3">Price</th>
            <th class="p-3">Actions</th>
            </tr>
        </thead>
        <tbody>
            <tr v-for="item in inventory" :key="item.id" class="border-t">
            <!-- Editing Mode -->
            <template v-if="editingId === item.id">
                <td class="p-3"><input v-model="editForm.name" class="border p-1 rounded w-full" /></td>
                <td class="p-3"><input v-model.number="editForm.current_stock" type="number" class="border p-1 rounded w-20" /></td>
                <td class="p-3 text-gray-600">{{ item.planned_consumption }}</td>
                <td class="p-3"><input v-model.number="editForm.price" type="number" step="0.01" class="border p-1 rounded w-20" /></td>
                <td class="p-3">
                    <div class="flex gap-1">
                    <button @click="saveEdit(item)" class="bg-green-500 text-white px-2 py-1 rounded text-sm hover:bg-green-600">Save</button>
                    <button @click="cancelEdit" class="bg-gray-300 px-2 py-1 rounded text-sm hover:bg-gray-400">Cancel</button>
                    </div>
                </td>
            </template>
            <!-- Display Mode -->
            <template v-else>
                <td class="p-3 font-medium">{{ item.name }}</td>
                <td class="p-3">
                    <div class="flex items-center gap-2">
                    <button @click="updateStock(item, -1)" class="bg-gray-200 px-2 rounded hover:bg-gray-300">-</button>
                    <span :class="{'text-red-500 font-bold': item.current_stock < item.planned_consumption}">{{ item.current_stock }}</span>
                    <button @click="updateStock(item, 1)" class="bg-gray-200 px-2 rounded hover:bg-gray-300">+</button>
                    </div>
                </td>
                <td class="p-3 text-gray-600">{{ item.planned_consumption }}</td>
                <td class="p-3">${{ item.price }}</td>
                <td class="p-3">
                    <div class="flex gap-1">
                    <button @click="startEdit(item)" class="bg-yellow-400 text-white px-2 py-1 rounded text-sm hover:bg-yellow-500">Edit</button>
                    <button @click="deleteItem(item)" class="bg-red-500 text-white px-2 py-1 rounded text-sm hover:bg-red-600">Delete</button>
                    </div>
                </td>
            </template>
            </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const inventory = ref([])
const newItem = ref({ name: '', current_stock: 0, price: 0 })
const editingId = ref(null)
const editForm = ref({ name: '', current_stock: 0, price: 0 })

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
    if (!newItem.value.name) return
    try {
        const res = await fetch('http://localhost:8080/api/inventory', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(newItem.value)
        })
        if (res.ok) {
            await fetchInventory()
            newItem.value = { name: '', current_stock: 0, price: 0 }
        }
    } catch (e) {
        console.error("Failed to add item", e)
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
    editForm.value = { name: item.name, current_stock: item.current_stock, price: item.price }
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
