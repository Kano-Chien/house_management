CREATE TABLE IF NOT EXISTS ingredients (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    current_stock REAL DEFAULT 0,
    unit TEXT,
    expiry_date TEXT, -- SQLite doesn't have DATE type, stores as TEXT (ISO8601)
    price REAL DEFAULT 0,
    category TEXT DEFAULT 'food',
    is_tracked BOOLEAN DEFAULT 1 -- SQLite uses 0/1 for boolean
);

CREATE TABLE IF NOT EXISTS recipes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    instructions TEXT,
    notes TEXT DEFAULT ''
);

CREATE TABLE IF NOT EXISTS recipe_ingredients (
    recipe_id INTEGER,
    ingredient_id INTEGER,
    quantity REAL NOT NULL,
    PRIMARY KEY (recipe_id, ingredient_id),
    FOREIGN KEY (recipe_id) REFERENCES recipes(id) ON DELETE CASCADE,
    FOREIGN KEY (ingredient_id) REFERENCES ingredients(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS meal_plan (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date TEXT NOT NULL,
    meal_type TEXT NOT NULL CHECK (meal_type IN ('Breakfast', 'Lunch', 'Dinner')),
    recipe_id INTEGER,
    is_cooked BOOLEAN DEFAULT 0,
    FOREIGN KEY (recipe_id) REFERENCES recipes(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS shopping_list_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    is_custom BOOLEAN DEFAULT 0,
    is_checked BOOLEAN DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
