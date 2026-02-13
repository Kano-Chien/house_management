<template>
  <div class="min-h-screen bg-gray-50 font-sans text-gray-900">
    <!-- Header / Nav -->
    <nav style="background-color: #E0E0E0;" class="text-gray-800 p-4 shadow-md">
      <div class="max-w-4xl mx-auto flex justify-between items-center">
        <h1 class="text-xl font-bold">🏠 Home Inventory</h1>
        <div class="flex gap-4">
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
          >Shopping List</button>
        </div>
      </div>
    </nav>

    <!-- Main Content -->
    <main class="max-w-4xl mx-auto mt-6">
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
