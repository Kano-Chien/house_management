package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Kano-Chien/house_management/backend/database"
	"github.com/Kano-Chien/house_management/backend/handlers"
	_ "modernc.org/sqlite"
)

func main() {
	loadEnv(".env") 

	// Database connection string (file path for SQLite)
	dbPath := os.Getenv("DATABASE_URL") //your database path put in .env file
	if dbPath == "" {
		dbPath = "house.db"
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal("Cannot connect to database:", err)
	}
	fmt.Println("Connected to the database successfully.")

	// Enable Foreign Keys and WAL mode for SQLite
	if _, err := db.Exec("PRAGMA foreign_keys = ON; PRAGMA journal_mode = WAL;"); err != nil {
		log.Fatal("Failed to enable foreign keys or WAL mode:", err)
	}

	// Avoid "database is locked" errors by limiting to one connection
	// This serializes all db access which is fine for this scale and Pi Zero
	db.SetMaxOpenConns(1)

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatal("Migration failed:", err)
	}
	fmt.Println("Database migrations applied successfully.")

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

	mux.HandleFunc("/api/inventory/stock-in", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			inventoryHandler.StockIn(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/inventory/batches", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			inventoryHandler.GetBatches(w, r)
		case "PUT":
			inventoryHandler.UpdateBatch(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/inventory/batches/delete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			inventoryHandler.DeleteBatch(w, r)
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

	mux.HandleFunc("/api/recipes/ingredients/replace", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			recipeHandler.ReplaceRecipeIngredient(w, r)
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

	mux.HandleFunc("/api/mealplan/ingredients", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			mealPlanHandler.GetMealPlanIngredients(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/mealplan/update", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			mealPlanHandler.UpdateMealPlan(w, r)
		} else {
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

	mux.HandleFunc("/api/mealplan/cook", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			mealPlanHandler.CookMeal(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/mealplan/cook-day", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			mealPlanHandler.CookAllDay(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/shopping-list", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			shoppingListHandler.GetShoppingList(w, r)
		case "POST":
			shoppingListHandler.AddCustomItem(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/shopping-list/check", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" {
			shoppingListHandler.ToggleItemChecked(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/shopping-list/delete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			shoppingListHandler.DeleteItem(w, r)
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

	mux.HandleFunc("/api/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			handlers.UploadImage(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Serve the static uploads directory
	uploadsDir := http.Dir("./uploads")
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(uploadsDir)))

	// Register SPA Handler
	spa := spaHandler{staticPath: "./dist", indexPath: "index.html"}
	mux.Handle("/", spa)

	handler := enableCORS(mux)

	// Start Server
	port := ":8080"
	fmt.Printf("Server starting on port %s\n", port)
	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatal(err)
	}
}

func enableCORS(next http.Handler) http.Handler {
	allowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		allowedOrigins = "*"
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if allowedOrigins == "*" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			for _, o := range strings.Split(allowedOrigins, ",") {
				if strings.TrimSpace(o) == origin {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					break
				}
			}
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// spaHandler implements the http.Handler interface, so we can use it
// to respond to HTTP requests. The path to the static directory and
// error helper logic are encapsulated in this connection.
type spaHandler struct {
	staticPath string
	indexPath  string
}

// ServeHTTP inspects the URL path to locate a file within the static dir
// on the SPA handler. If a file is found, it will be served. If not, the
// file located at the index path on the SPA handler will be served.
func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// join internally call path.Clean to prevent directory traversal
	path := filepath.Join(h.staticPath, r.URL.Path)

	// check whether a file exists or is a directory at the given path
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// file does not exist, serve index.html (SPA fallback)
			http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
			return
		}
		// Other error (e.g. permission denied), return 500
		log.Printf("Error stating file %s: %v", path, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if fi.IsDir() {
		// directory, serve index.html (SPA fallback)
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	}

	// otherwise, use http.FileServer to serve the static file
	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}
func loadEnv(path string) {
	fmt.Printf("Loading .env from: %s\n", path) // fmt for print
	f, err := os.Open(path)
	if err != nil {  // nil is a special value in Go that represents the absence of a value
		fmt.Printf("Error open .env: %v\n", err)
		return
	}
	defer f.Close() // defer is a keyword in Go that defers the execution of a function until the surrounding function returns
	scanner := bufio.NewScanner(f) // bufio.NewScanner is a function that creates a new scanner that reads from the given reader
	for scanner.Scan() { // scanner.Scan() is a method that reads the next token from the scanner
		line := strings.TrimSpace(scanner.Text()) // strings.TrimSpace is a function that removes leading and trailing whitespace from a string
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			os.Setenv(parts[0], parts[1]) // os.Setenv is a function that sets an environment variable
			fmt.Printf("Loaded env: %s\n", parts[0])
		}
	}
}
