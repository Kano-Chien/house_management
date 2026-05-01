<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-4">
      <h2 class="text-2xl font-bold">Inventory</h2>
      <div class="flex gap-2">
        <button @click="showStockInModal = true"
          class="flex items-center gap-1.5 px-4 py-2 rounded-lg text-sm font-medium bg-green-500 text-white hover:bg-green-600 shadow-sm hover:shadow transition-all duration-200">
          + Stock In
        </button>
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
    </div>

    <!-- Loading skeleton -->
    <div v-if="loading" class="space-y-3 mt-4">
      <div v-for="i in 4" :key="i" class="h-16 bg-gray-100 rounded-2xl animate-pulse" />
    </div>

    <template v-else>
    <!-- Category Filter -->
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
              class="border border-gray-200 p-2 rounded-lg focus:ring-2 focus:ring-blue-400 focus:border-blue-400 focus:outline-none transition-all w-full md:w-40" />
          </div>
          <div class="flex flex-col gap-1">
            <label class="text-xs font-medium text-gray-500 uppercase tracking-wide">Stock</label>
            <input v-model.number="newItem.current_stock" type="number" placeholder="0"
              @keyup.enter="addItem"
              class="border border-gray-200 p-2 rounded-lg focus:ring-2 focus:ring-blue-400 focus:border-blue-400 focus:outline-none transition-all w-full md:w-20" />
          </div>
          <div class="flex flex-col gap-1">
            <label class="text-xs font-medium text-gray-500 uppercase tracking-wide">Price</label>
            <input v-model.number="newItem.price" type="number" step="1" min="0" placeholder="$0"
              @keyup.enter="addItem"
              class="border border-gray-200 p-2 rounded-lg focus:ring-2 focus:ring-blue-400 focus:border-blue-400 focus:outline-none transition-all w-full md:w-20" />
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
              'px-5 py-2 rounded-lg text-sm font-medium transition-all duration-200 w-full md:w-auto',
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
    <div class="bg-white rounded-xl shadow overflow-x-auto">
      <table class="w-full text-left">
        <thead class="bg-gray-50">
          <tr>
            <th class="p-3 text-xs font-semibold text-gray-500 uppercase tracking-wide">
              <button @click="toggleSort('name')" class="flex items-center gap-1 hover:text-gray-700 transition-colors">
                Name <SortIcon :active="sortBy === 'name'" :dir="sortDir" />
              </button>
            </th>
            <th class="p-3 text-xs font-semibold text-gray-500 uppercase tracking-wide">Category</th>
            <th class="p-3 text-xs font-semibold text-gray-500 uppercase tracking-wide">Stock</th>
            <th class="p-3 text-xs font-semibold text-gray-500 uppercase tracking-wide">Consume</th>
            <th class="p-3 text-xs font-semibold text-gray-500 uppercase tracking-wide">
              <button @click="toggleSort('expiry')" class="flex items-center gap-1 hover:text-gray-700 transition-colors">
                Expiry <SortIcon :active="sortBy === 'expiry'" :dir="sortDir" />
              </button>
            </th>
            <th class="p-3 text-xs font-semibold text-gray-500 uppercase tracking-wide">Price</th>
            <th class="p-3 text-xs font-semibold text-gray-500 uppercase tracking-wide">Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in sortedInventory" :key="item.id"
            :class="['border-t transition-colors duration-150', justAdded === item.id ? 'bg-green-50' : 'hover:bg-gray-50']">
            <td class="p-3 font-medium">{{ item.name }}</td>
            <td class="p-3">
              <span :class="[
                'inline-block px-2 py-0.5 rounded-full text-xs font-medium',
                item.category === 'food' ? 'bg-orange-100 text-orange-700' : 'bg-blue-100 text-blue-700'
              ]">{{ item.category === 'food' ? '🍖 Food' : '🧴 Daily' }}</span>
            </td>
            <td class="p-3">
              <span :class="{'text-red-500 font-bold': item.current_stock < item.planned_consumption}">
                {{ item.current_stock }}
              </span>
            </td>
            <td class="p-3 text-gray-500">{{ item.planned_consumption }}</td>
            <td class="p-3">
              <span v-if="item.earliest_expiry" :class="[
                'inline-block px-2 py-0.5 rounded text-xs font-medium',
                isExpired(item.earliest_expiry)
                  ? 'bg-red-100 text-red-700'
                  : isExpiringSoon(item.earliest_expiry)
                    ? 'bg-yellow-100 text-yellow-700'
                    : 'text-gray-600'
              ]">
                {{ isExpired(item.earliest_expiry) || isExpiringSoon(item.earliest_expiry) ? '⚠ ' : '' }}{{ formatExpiry(item.earliest_expiry) }}
              </span>
              <span v-else class="text-gray-300">—</span>
            </td>
            <td class="p-3">${{ item.price }}</td>
            <td class="p-3">
              <div class="flex gap-1">
                <button @click="startEdit(item)" class="bg-yellow-400 text-white px-2 py-1 rounded-lg text-sm hover:bg-yellow-500 transition-colors">Edit</button>
                <button @click="deleteItem(item)" class="bg-red-500 text-white px-2 py-1 rounded-lg text-sm hover:bg-red-600 transition-colors">Delete</button>
              </div>
            </td>
          </tr>
          <tr v-if="sortedInventory.length === 0">
            <td colspan="7" class="p-8 text-center text-gray-400">
              <div class="text-3xl mb-2">📦</div>
              No items in this category
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Edit Modal -->
    <Transition name="fade">
      <div v-if="showEditModal" class="fixed inset-0 bg-black/40 z-50 flex items-center justify-center p-4" @click.self="cancelEdit">
        <div class="bg-white rounded-2xl shadow-xl w-full max-w-md flex flex-col max-h-[90vh]">

          <!-- Header -->
          <div class="p-5 border-b flex items-center justify-between">
            <h3 class="text-lg font-semibold">Edit — {{ editingItem?.name }}</h3>
            <button @click="cancelEdit" class="text-gray-400 hover:text-gray-600 text-xl leading-none">✕</button>
          </div>

          <!-- Fields -->
          <div class="p-5 space-y-4">
            <div class="grid grid-cols-2 gap-3">
              <div class="flex flex-col gap-1">
                <label class="text-xs font-medium text-gray-500 uppercase tracking-wide">Name</label>
                <input v-model="editForm.name"
                  class="border border-gray-200 p-2 rounded-lg focus:ring-2 focus:ring-blue-400 focus:outline-none text-sm" />
              </div>
              <div class="flex flex-col gap-1">
                <label class="text-xs font-medium text-gray-500 uppercase tracking-wide">Category</label>
                <select v-model="editForm.category"
                  class="border border-gray-200 p-2 rounded-lg bg-white text-sm focus:ring-2 focus:ring-blue-400 focus:outline-none">
                  <option value="food">🍖 Food</option>
                  <option value="daily">🧴 Daily</option>
                </select>
              </div>
              <div class="flex flex-col gap-1">
                <label class="text-xs font-medium text-gray-500 uppercase tracking-wide">Price</label>
                <input v-model.number="editForm.price" type="number" step="1" min="0"
                  class="border border-gray-200 p-2 rounded-lg focus:ring-2 focus:ring-blue-400 focus:outline-none text-sm" />
              </div>
              <div class="flex flex-col gap-1 justify-end">
                <label class="flex items-center gap-2 text-sm text-gray-600 cursor-pointer pb-2">
                  <input type="checkbox" v-model="editForm.is_tracked" class="rounded text-blue-500 focus:ring-blue-400" />
                  Track stock level
                </label>
              </div>
            </div>

            <!-- Batches section -->
            <div class="border-t pt-4">
              <div class="text-xs font-semibold text-gray-400 uppercase tracking-wide mb-3">Batches</div>
              <div v-if="!editBatches || editBatches.length === 0" class="text-sm text-gray-400 italic">No batches recorded.</div>
              <table v-else class="w-full text-sm">
                <thead>
                  <tr class="text-xs text-gray-400">
                    <th class="text-left pb-2 pr-3 font-medium">Qty</th>
                    <th class="text-left pb-2 pr-3 font-medium">Expiry</th>
                    <th class="pb-2"></th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="batch in editBatches" :key="batch.id" class="border-t border-gray-100">
                    <template v-if="batchEditId === batch.id">
                      <td class="py-2 pr-3">
                        <input v-model.number="batchEditForm.quantity" type="number" min="0"
                          class="border p-1 rounded w-16 text-sm focus:ring-2 focus:ring-blue-400 focus:outline-none" />
                      </td>
                      <td class="py-2 pr-3">
                        <input v-model="batchEditForm.expiry_date" type="date"
                          class="border p-1 rounded text-sm focus:ring-2 focus:ring-blue-400 focus:outline-none" />
                      </td>
                      <td class="py-2">
                        <div class="flex gap-1">
                          <button @click="saveBatchEdit" class="bg-green-500 text-white px-2 py-0.5 rounded text-xs hover:bg-green-600">Save</button>
                          <button @click="batchEditId = null" class="bg-gray-200 text-gray-600 px-2 py-0.5 rounded text-xs hover:bg-gray-300">Cancel</button>
                        </div>
                      </td>
                    </template>
                    <template v-else>
                      <td class="py-2 pr-3 font-medium">{{ batch.quantity }}</td>
                      <td class="py-2 pr-3">
                        <span v-if="batch.expiry_date" :class="[
                          'px-1.5 py-0.5 rounded text-xs font-medium',
                          isExpired(batch.expiry_date)
                            ? 'bg-red-100 text-red-700'
                            : isExpiringSoon(batch.expiry_date)
                              ? 'bg-yellow-100 text-yellow-700'
                              : 'text-gray-600'
                        ]">{{ isExpired(batch.expiry_date) || isExpiringSoon(batch.expiry_date) ? '⚠ ' : '' }}{{ formatExpiry(batch.expiry_date) }}</span>
                        <span v-else class="text-gray-300">No expiry</span>
                      </td>
                      <td class="py-2">
                        <div class="flex gap-1">
                          <button @click="startBatchEdit(batch)" class="text-yellow-500 hover:text-yellow-600 text-xs font-medium">Edit</button>
                          <button @click="deleteBatch(batch)" class="text-red-400 hover:text-red-600 text-xs font-medium">Delete</button>
                        </div>
                      </td>
                    </template>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- Footer -->
          <div class="p-5 border-t flex justify-end gap-2">
            <button @click="cancelEdit"
              class="px-4 py-2 rounded-lg text-sm text-gray-600 bg-gray-100 hover:bg-gray-200 transition-colors">
              Cancel
            </button>
            <button @click="saveEdit"
              class="px-5 py-2 rounded-lg text-sm font-medium bg-blue-500 text-white hover:bg-blue-600 shadow-sm transition-colors">
              Save
            </button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Stock In Modal -->
    <Transition name="fade">
      <div v-if="showStockInModal" class="fixed inset-0 bg-black/40 z-50 flex items-center justify-center p-4" @click.self="showStockInModal = false">
        <div class="bg-white rounded-2xl shadow-xl w-full max-w-lg flex flex-col max-h-[90vh]">
          <div class="p-5 border-b flex items-center justify-between">
            <h3 class="text-lg font-semibold">Stock In</h3>
            <button @click="showStockInModal = false" class="text-gray-400 hover:text-gray-600 text-xl leading-none">✕</button>
          </div>
          <div class="px-5 pt-4 grid grid-cols-[1fr_64px_120px_28px] gap-2 text-xs font-medium text-gray-400 uppercase tracking-wide">
            <span>Item</span><span>Qty</span><span>Expiry (optional)</span><span></span>
          </div>
          <div class="px-5 py-3 space-y-2 overflow-y-auto flex-1">
            <div v-for="(row, idx) in stockInRows" :key="idx"
              class="grid grid-cols-[1fr_64px_120px_28px] gap-2 items-center">
              <select v-model="row.ingredient_id"
                class="border border-gray-200 p-2 rounded-lg bg-white text-sm focus:ring-2 focus:ring-green-400 focus:outline-none">
                <option value="" disabled>Select item</option>
                <option v-for="item in inventory" :key="item.id" :value="item.id">{{ item.name }}</option>
              </select>
              <input v-model.number="row.quantity" type="number" min="1" placeholder="1"
                class="border border-gray-200 p-2 rounded-lg text-sm focus:ring-2 focus:ring-green-400 focus:outline-none w-full" />
              <input v-model="row.expiry_date" type="date"
                class="border border-gray-200 p-2 rounded-lg text-sm focus:ring-2 focus:ring-green-400 focus:outline-none w-full" />
              <button @click="stockInRows.splice(idx, 1)"
                class="text-gray-300 hover:text-red-400 transition-colors text-lg leading-none font-medium">✕</button>
            </div>
          </div>
          <div class="p-5 border-t flex items-center justify-between gap-3">
            <button @click="addStockInRow" class="text-sm text-green-600 hover:text-green-700 font-medium transition-colors">+ Add row</button>
            <div class="flex gap-2">
              <button @click="showStockInModal = false"
                class="px-4 py-2 rounded-lg text-sm text-gray-600 bg-gray-100 hover:bg-gray-200 transition-colors">Cancel</button>
              <button @click="submitStockIn" :disabled="stockInSubmitting || !stockInValid"
                :class="[
                  'px-5 py-2 rounded-lg text-sm font-medium transition-all duration-200',
                  stockInSubmitting ? 'bg-green-500 text-white scale-95'
                    : !stockInValid ? 'bg-gray-200 text-gray-400 cursor-not-allowed'
                    : 'bg-green-500 text-white hover:bg-green-600 shadow-sm'
                ]">
                {{ stockInSubmitting ? '✓ Done!' : 'Confirm' }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </Transition>
    </template>

    <!-- Confirm delete item -->
    <ConfirmDialog
      v-if="confirmDelete"
      :message="`Delete &quot;${confirmDelete.name}&quot;? This cannot be undone.`"
      @confirm="confirmDeleteItem"
      @cancel="confirmDelete = null"
    />

    <!-- Confirm delete batch -->
    <ConfirmDialog
      v-if="confirmDeleteBatch"
      title="Delete batch?"
      :message="`Remove batch of ${confirmDeleteBatch.quantity}? This cannot be undone.`"
      @confirm="confirmDeleteBatchItem"
      @cancel="confirmDeleteBatch = null"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick, watch } from 'vue'
import ConfirmDialog from './ui/ConfirmDialog.vue'
import { useToast } from '../composables/useToast.js'

const { toast } = useToast()

const SortIcon = {
  props: ['active', 'dir'],
  template: `<span class="inline-flex flex-col leading-none opacity-40 text-[10px]" :class="{ 'opacity-100 text-blue-500': active }">
    <span :class="{ 'opacity-30': active && dir === 'desc' }">▲</span>
    <span :class="{ 'opacity-30': active && dir === 'asc' }">▼</span>
  </span>`
}

const loading = ref(true)
const inventory = ref([])
const newItem = ref({ name: '', current_stock: 0, price: 0, category: 'food', is_tracked: true })
const showAddForm = ref(false)
const adding = ref(false)
const justAdded = ref(null)
const nameInputRef = ref(null)

// Edit modal
const showEditModal = ref(false)
const editingItem = ref(null)
const editForm = ref({ name: '', price: 0, category: 'food', is_tracked: true })
const editBatches = ref([])
const batchEditId = ref(null)
const batchEditForm = ref({ quantity: 1, expiry_date: '' })

const confirmDelete = ref(null)
const confirmDeleteBatch = ref(null)

const startEdit = async (item) => {
  editingItem.value = item
  editForm.value = { name: item.name, price: item.price, category: item.category || 'food', is_tracked: item.is_tracked }
  editBatches.value = []
  batchEditId.value = null
  showEditModal.value = true
  await loadBatches(item.id)
}

const cancelEdit = () => {
  showEditModal.value = false
  editingItem.value = null
  batchEditId.value = null
}

const saveEdit = async () => {
  try {
    const res = await fetch('/api/inventory/edit', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id: editingItem.value.id, ...editForm.value })
    })
    if (res.ok) {
      await fetchInventory()
      showEditModal.value = false
      editingItem.value = null
    }
  } catch (e) {
    toast('Failed to save changes', 'error')
  }
}

const loadBatches = async (ingredientId) => {
  try {
    const res = await fetch(`/api/inventory/batches?ingredient_id=${ingredientId}`)
    if (res.ok) editBatches.value = (await res.json()) || []
  } catch (e) {
    toast('Failed to load batches', 'error')
  }
}

const startBatchEdit = (batch) => {
  batchEditId.value = batch.id
  batchEditForm.value = { quantity: batch.quantity, expiry_date: batch.expiry_date || '' }
}

const saveBatchEdit = async () => {
  try {
    const res = await fetch('/api/inventory/batches', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id: batchEditId.value, ...batchEditForm.value })
    })
    if (res.ok) {
      batchEditId.value = null
      await loadBatches(editingItem.value.id)
      await fetchInventory()
    }
  } catch (e) {
    toast('Failed to save batch', 'error')
  }
}

const deleteBatch = (batch) => {
  confirmDeleteBatch.value = batch
}
const confirmDeleteBatchItem = async () => {
  const batch = confirmDeleteBatch.value
  confirmDeleteBatch.value = null
  try {
    const res = await fetch('/api/inventory/batches/delete', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id: batch.id })
    })
    if (res.ok) {
      await loadBatches(editingItem.value.id)
      await fetchInventory()
    } else {
      toast('Failed to delete batch', 'error')
    }
  } catch (e) {
    toast('Failed to delete batch', 'error')
  }
}

// Sorting
const sortBy = ref('name')
const sortDir = ref('asc')
const toggleSort = (field) => {
  if (sortBy.value === field) {
    sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortBy.value = field
    sortDir.value = 'asc'
  }
}

// Stock In modal
const showStockInModal = ref(false)
const stockInRows = ref([])
const stockInSubmitting = ref(false)
const addStockInRow = () => stockInRows.value.push({ ingredient_id: '', quantity: 1, expiry_date: '' })
const stockInValid = computed(() =>
  stockInRows.value.length > 0 && stockInRows.value.every(r => r.ingredient_id !== '' && r.quantity > 0)
)
watch(showStockInModal, (val) => {
  if (val) stockInRows.value = [{ ingredient_id: '', quantity: 1, expiry_date: '' }]
})
const submitStockIn = async () => {
  if (!stockInValid.value || stockInSubmitting.value) return
  stockInSubmitting.value = true
  try {
    const res = await fetch('/api/inventory/stock-in', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ items: stockInRows.value })
    })
    if (res.ok) {
      await fetchInventory()
      setTimeout(() => { showStockInModal.value = false; stockInSubmitting.value = false }, 600)
    }
  } catch (e) {
    toast('Failed to stock in', 'error')
    stockInSubmitting.value = false
  }
}

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
const sortedInventory = computed(() => {
  const items = [...filteredInventory.value]
  const dir = sortDir.value === 'asc' ? 1 : -1
  return items.sort((a, b) => {
    if (sortBy.value === 'expiry') {
      if (!a.earliest_expiry && !b.earliest_expiry) return 0
      if (!a.earliest_expiry) return 1
      if (!b.earliest_expiry) return -1
      return a.earliest_expiry.localeCompare(b.earliest_expiry) * dir
    }
    return a.name.localeCompare(b.name) * dir
  })
})

// Expiry helpers
const today = () => { const d = new Date(); d.setHours(0, 0, 0, 0); return d }
const isExpiringSoon = (d) => { if (!d) return false; const diff = (new Date(d + 'T00:00:00') - today()) / 86400000; return diff >= 0 && diff <= 3 }
const isExpired = (d) => { if (!d) return false; return new Date(d + 'T00:00:00') < today() }
const formatExpiry = (d) => { if (!d) return '—'; const [, m, day] = d.split('-'); return `${parseInt(m)}/${parseInt(day)}` }

// Auto-focus name input
watch(showAddForm, async (val) => { if (val) { await nextTick(); nameInputRef.value?.focus() } })

const fetchInventory = async () => {
  try {
    const res = await fetch('/api/inventory')
    if (res.ok) inventory.value = await res.json()
  } catch (e) {
    toast('Failed to load inventory', 'error')
  } finally {
    loading.value = false
  }
}

const addItem = async () => {
  if (!newItem.value.name || adding.value) return
  adding.value = true
  try {
    const res = await fetch('/api/inventory', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(newItem.value)
    })
    if (res.ok) {
      const added = await res.json()
      await fetchInventory()
      justAdded.value = added.id
      setTimeout(() => { justAdded.value = null }, 1500)
      newItem.value = { name: '', current_stock: 0, price: 0, category: newItem.value.category, is_tracked: true }
      await nextTick()
      nameInputRef.value?.focus()
    }
  } catch (e) {
    toast('Failed to add item', 'error')
  } finally {
    setTimeout(() => { adding.value = false }, 600)
  }
}

const deleteItem = (item) => {
  confirmDelete.value = item
}
const confirmDeleteItem = async () => {
  const item = confirmDelete.value
  confirmDelete.value = null
  try {
    const res = await fetch('/api/inventory/delete', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id: item.id }),
    })
    if (res.ok) {
      await fetchInventory()
      toast(`"${item.name}" deleted`)
    } else {
      toast('Failed to delete item', 'error')
    }
  } catch (e) {
    toast('Failed to delete item', 'error')
  }
}

onMounted(fetchInventory)
</script>

<style scoped>
.slide-enter-active, .slide-leave-active {
  transition: all 0.25s ease;
  overflow: hidden;
}
.slide-enter-from, .slide-leave-to {
  opacity: 0; max-height: 0; padding-top: 0; padding-bottom: 0; margin-bottom: 0;
}
.slide-enter-to, .slide-leave-from {
  opacity: 1; max-height: 200px;
}
.fade-enter-active, .fade-leave-active { transition: opacity 0.2s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
