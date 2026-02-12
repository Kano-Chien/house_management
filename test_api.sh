#!/bin/bash

BASE_URL="http://localhost:8080/api"

echo "Testing Inventory API..."
# Add Ingredient
curl -X POST $BASE_URL/inventory -d '{"name":"Test Item","current_stock":10,"unit":"kg"}' -H "Content-Type: application/json"
echo -e "\nAdded Item"

# Get Inventory
curl $BASE_URL/inventory
echo -e "\nFetched Inventory"

echo -e "\n-------------------"
echo "Testing Recipe API..."
# Create Recipe
curl -X POST $BASE_URL/recipes -d '{"name":"Test Recipe","instructions":"Mix things"}' -H "Content-Type: application/json"
echo -e "\nCreated Recipe"

# Get Recipes
curl $BASE_URL/recipes
echo -e "\nFetched Recipes"

echo -e "\n-------------------"
echo "Testing Meal Plan API..."
# Schedule Meal (Assuming Recipe ID 1 exists from above or previous runs)
TODAY=$(date +%Y-%m-%d)
curl -X POST $BASE_URL/mealplan -d "{\"date\":\"$TODAY\",\"meal_type\":\"Lunch\",\"recipe_id\":1}" -H "Content-Type: application/json"
echo -e "\nScheduled Meal"

# Get Meal Plan
curl $BASE_URL/mealplan
echo -e "\nFetched Meal Plan"
