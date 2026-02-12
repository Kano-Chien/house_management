<template>
  <div class="min-h-screen bg-gray-50 font-sans text-gray-900">
    <!-- Header / Nav -->
    <nav class="bg-blue-600 text-white p-4 shadow-md">
      <div class="max-w-4xl mx-auto flex justify-between items-center">
        <h1 class="text-xl font-bold">🏠 Home Inventory</h1>
        <div class="flex gap-4">
          <button 
            @click="currentTab = 'inventory'"
            :class="['px-3 py-1 rounded', currentTab === 'inventory' ? 'bg-blue-800' : 'hover:bg-blue-500']"
          >Inventory</button>
          <button 
            @click="currentTab = 'recipes'"
             :class="['px-3 py-1 rounded', currentTab === 'recipes' ? 'bg-blue-800' : 'hover:bg-blue-500']"
          >Recipes</button>
          <button 
            @click="currentTab = 'planning'"
             :class="['px-3 py-1 rounded', currentTab === 'planning' ? 'bg-blue-800' : 'hover:bg-blue-500']"
          >Meal Plan</button>
          <button 
            @click="currentTab = 'shopping'"
             :class="['px-3 py-1 rounded', currentTab === 'shopping' ? 'bg-blue-800' : 'hover:bg-blue-500']"
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
