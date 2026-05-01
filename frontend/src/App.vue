<template>
  <div class="min-h-screen bg-gray-50 font-sans text-gray-900">
    <!-- Toast layer (global) -->
    <AppToast />

    <!-- Header -->
    <nav class="bg-white border-b border-gray-200 sticky top-0 z-40 shadow-sm">
      <div class="max-w-4xl mx-auto px-4 h-14 flex items-center justify-between">
        <h1 class="text-base font-bold text-gray-900">🏠 Home</h1>
        <!-- Desktop nav -->
        <div class="hidden md:flex gap-1 bg-gray-100 rounded-xl p-1">
          <button
            v-for="tab in tabs" :key="tab.id"
            @click="currentTab = tab.id"
            :class="[
              'px-4 py-1.5 rounded-lg text-sm font-medium transition-all',
              currentTab === tab.id
                ? 'bg-white text-primary-600 shadow-sm'
                : 'text-gray-500 hover:text-gray-700'
            ]"
          >{{ tab.label }}</button>
        </div>
      </div>
    </nav>

    <!-- Mobile bottom nav -->
    <nav class="md:hidden fixed bottom-0 inset-x-0 bg-white border-t border-gray-200 flex z-40 shadow-[0_-2px_10px_rgba(0,0,0,0.05)]">
      <button
        v-for="tab in tabs" :key="tab.id"
        @click="currentTab = tab.id"
        :class="[
          'flex flex-col items-center justify-center gap-1 flex-1 py-2 transition-colors',
          currentTab === tab.id ? 'text-primary-600' : 'text-gray-400'
        ]"
      >
        <span class="text-xl leading-none">{{ tab.icon }}</span>
        <span class="text-[10px] font-medium">{{ tab.label }}</span>
        <span
          v-if="currentTab === tab.id"
          class="w-1 h-1 rounded-full bg-primary-500"
        />
      </button>
    </nav>

    <!-- Main content — v-show keeps components mounted between tab switches -->
    <main class="max-w-4xl mx-auto pb-24 md:pb-8 px-0">
      <InventoryTable v-show="currentTab === 'inventory'" />
      <RecipeManager  v-show="currentTab === 'recipes'"   />
      <MealPlanner    v-show="currentTab === 'planning'"  />
      <ShoppingList   v-show="currentTab === 'shopping'"  />
    </main>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import InventoryTable from './components/InventoryTable.vue'
import RecipeManager  from './components/RecipeManager.vue'
import MealPlanner    from './components/MealPlanner.vue'
import ShoppingList   from './components/ShoppingList.vue'
import AppToast       from './components/ui/AppToast.vue'

const tabs = [
  { id: 'inventory', label: 'Inventory', icon: '📦' },
  { id: 'recipes',   label: 'Recipes',   icon: '📖' },
  { id: 'planning',  label: 'Meal Plan', icon: '📅' },
  { id: 'shopping',  label: 'Shopping',  icon: '🛒' },
]

const savedTab = localStorage.getItem('house_app_tab')
const currentTab = ref(savedTab || 'inventory')
watch(currentTab, t => localStorage.setItem('house_app_tab', t))
</script>
