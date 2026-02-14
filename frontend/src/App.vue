<template>
  <div class="min-h-screen bg-gray-50 font-sans text-gray-900">
    <!-- Header / Nav -->
    <nav style="background-color: #E0E0E0;" class="text-gray-800 p-4 shadow-md sticky top-0 z-50">
      <div class="max-w-4xl mx-auto flex justify-between items-center">
        <h1 class="text-xl font-bold">🏠 Home Inventory</h1>
        <!-- Desktop Nav -->
        <div class="hidden md:flex gap-4">
          <button 
            @click="currentTab = 'inventory'"
            :class="['px-3 py-1 rounded font-medium transition-colors', currentTab === 'inventory' ? 'bg-gray-700 text-white' : 'hover:bg-gray-300']"
          >Inventory</button>
          <button 
            @click="currentTab = 'recipes'"
             :class="['px-3 py-1 rounded font-medium transition-colors', currentTab === 'recipes' ? 'bg-gray-700 text-white' : 'hover:bg-gray-300']"
          >Recipes</button>
          <button 
            @click="currentTab = 'planning'"
             :class="['px-3 py-1 rounded font-medium transition-colors', currentTab === 'planning' ? 'bg-gray-700 text-white' : 'hover:bg-gray-300']"
          >Meal Plan</button>
          <button 
            @click="currentTab = 'shopping'"
             :class="['px-3 py-1 rounded font-medium transition-colors', currentTab === 'shopping' ? 'bg-gray-700 text-white' : 'hover:bg-gray-300']"
          >Shopping</button>
        </div>
      </div>
    </nav>

    <!-- Mobile Bottom Nav -->
    <nav class="md:hidden fixed bottom-0 left-0 right-0 bg-white border-t border-gray-200 flex justify-around p-2 z-50 shadow-[0_-2px_10px_rgba(0,0,0,0.05)]">
      <button 
        @click="currentTab = 'inventory'"
        :class="['flex flex-col items-center p-2 rounded-lg transition-colors w-full', currentTab === 'inventory' ? 'text-blue-600' : 'text-gray-400 hover:text-gray-600']"
      >
        <span class="text-xl leading-none">📦</span>
        <span class="text-[10px] font-medium mt-1">Inventory</span>
      </button>
      <button 
        @click="currentTab = 'recipes'"
        :class="['flex flex-col items-center p-2 rounded-lg transition-colors w-full', currentTab === 'recipes' ? 'text-blue-600' : 'text-gray-400 hover:text-gray-600']"
      >
        <span class="text-xl leading-none">📖</span>
        <span class="text-[10px] font-medium mt-1">Recipes</span>
      </button>
      <button 
        @click="currentTab = 'planning'"
        :class="['flex flex-col items-center p-2 rounded-lg transition-colors w-full', currentTab === 'planning' ? 'text-blue-600' : 'text-gray-400 hover:text-gray-600']"
      >
        <span class="text-xl leading-none">📅</span>
        <span class="text-[10px] font-medium mt-1">Plan</span>
      </button>
      <button 
        @click="currentTab = 'shopping'"
        :class="['flex flex-col items-center p-2 rounded-lg transition-colors w-full', currentTab === 'shopping' ? 'text-blue-600' : 'text-gray-400 hover:text-gray-600']"
      >
        <span class="text-xl leading-none">🛒</span>
        <span class="text-[10px] font-medium mt-1">Shop</span>
      </button>
    </nav>

    <!-- Main Content -->
    <main class="max-w-4xl mx-auto mt-6 pb-24 md:pb-6">
      <InventoryTable v-if="currentTab === 'inventory'" />
      <RecipeManager v-if="currentTab === 'recipes'" />
      <MealPlanner v-if="currentTab === 'planning'" />
      <ShoppingList v-if="currentTab === 'shopping'" />
    </main>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import InventoryTable from './components/InventoryTable.vue'
import RecipeManager from './components/RecipeManager.vue'
import MealPlanner from './components/MealPlanner.vue'
import ShoppingList from './components/ShoppingList.vue'

const currentTab = ref('inventory')
</script>

<style>
</style>
