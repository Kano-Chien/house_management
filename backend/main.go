package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Kano-Chien/house_management/backend/handlers"
	_ "github.com/lib/pq"
)

func main() {
	loadEnv(".env")

	// Database connection string
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "user=house_user password=house_pass dbname=house_management sslmode=disable"
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal("Cannot connect to database:", err)
	}
	fmt.Println("Connected to the database successfully.")

	// Execute Schema
	schemaContent, err := os.ReadFile("database/schema.sql")
	if err != nil {
		log.Fatal("Error reading schema file:", err)
	}
	if _, err := db.Exec(string(schemaContent)); err != nil {
		log.Fatal("Error executing schema:", err)
	}
	fmt.Println("Database schema applied successfully.")

	// Initialize Handlers
	inventoryHandler := &handlers.InventoryHandler{DB: db}
	recipeHandler := &handlers.RecipeHandler{DB: db}
	mealPlanHandler := &handlers.MealPlanHandler{DB: db}
	shoppingListHandler := &handlers.ShoppingListHandler{DB: db}
	lineNotifyHandler := &handlers.LineNotifyHandler{DB: db}

	// Router setup - using path-only patterns with method checks
	mux := http.NewServeMux()

	mux.HandleFunc("/api/inventory", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			inventoryHandler.GetInventory(w, r)
		case "POST":
			inventoryHandler.AddIngredient(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/inventory/stock", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" {
			inventoryHandler.UpdateStock(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/inventory/edit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" {
			inventoryHandler.EditIngredient(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/inventory/delete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			inventoryHandler.DeleteIngredient(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/recipes", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			recipeHandler.GetRecipes(w, r)
		case "POST":
			recipeHandler.CreateRecipe(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/recipes/ingredients", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			recipeHandler.GetRecipeIngredients(w, r)
		case "POST":
			recipeHandler.AddRecipeIngredient(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/recipes/ingredients/remove", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			recipeHandler.RemoveRecipeIngredient(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/recipes/delete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			recipeHandler.DeleteRecipe(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/recipes/edit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" {
			recipeHandler.UpdateRecipeName(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/recipes/ingredients/edit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" {
			recipeHandler.UpdateIngredientQuantity(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/mealplan", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			mealPlanHandler.GetMealPlan(w, r)
		case "POST":
			mealPlanHandler.ScheduleMeal(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/mealplan/delete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			mealPlanHandler.DeleteMealPlan(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/shopping-list", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			shoppingListHandler.GetShoppingList(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// CORS Middleware
	mux.HandleFunc("/api/line/send-shopping-list", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			lineNotifyHandler.SendShoppingList(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	handler := enableCORS(mux)

	// Start Server
	port := ":8080"
	fmt.Printf("Server starting on port %s\n", port)
	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatal(err)
	}
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
func loadEnv(path string) {
	fmt.Printf("Loading .env from: %s\n", path)
	f, err := os.Open(path)
	if err != nil {
		fmt.Printf("Error open .env: %v\n", err)
		return
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			os.Setenv(parts[0], parts[1])
			fmt.Printf("Loaded env: %s\n", parts[0])
		}
	}
}
