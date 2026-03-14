CREATE TABLE IF NOT EXISTS ingredients (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    current_stock INTEGER DEFAULT 0,
    min_stock INTEGER DEFAULT 3, -- Minimum stock threshold for shopping list alerts
    price INTEGER DEFAULT 0,
    category TEXT DEFAULT 'food',
    is_tracked BOOLEAN DEFAULT 1 -- SQLite uses 0/1 for boolean
);

CREATE TABLE IF NOT EXISTS recipes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    notes TEXT DEFAULT ''
);

CREATE TABLE IF NOT EXISTS recipe_ingredients (
    recipe_id INTEGER,
    ingredient_id INTEGER,
    quantity INTEGER NOT NULL,
    PRIMARY KEY (recipe_id, ingredient_id),
    FOREIGN KEY (recipe_id) REFERENCES recipes(id) ON DELETE CASCADE,
    FOREIGN KEY (ingredient_id) REFERENCES ingredients(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS meal_plan (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date TEXT NOT NULL,
    meal_type TEXT NOT NULL CHECK (meal_type IN ('Breakfast', 'Lunch', 'Dinner')),
    recipe_id INTEGER,
    custom_name TEXT, -- For meals without a recipe
    is_cooked BOOLEAN DEFAULT 0,
    FOREIGN KEY (recipe_id) REFERENCES recipes(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS shopping_list_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    ingredient_id INTEGER, -- FK to ingredients for auto-generated items
    is_custom BOOLEAN DEFAULT 0,
    is_checked BOOLEAN DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (ingredient_id) REFERENCES ingredients(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS ingredient_batches (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ingredient_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    expiry_date TEXT, -- YYYY-MM-DD, nullable
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (ingredient_id) REFERENCES ingredients(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_meal_plan_date ON meal_plan(date);
CREATE INDEX IF NOT EXISTS idx_meal_plan_recipe ON meal_plan(recipe_id);
CREATE INDEX IF NOT EXISTS idx_ingredients_name ON ingredients(name);
CREATE INDEX IF NOT EXISTS idx_shopping_ingredient ON shopping_list_items(ingredient_id);
CREATE INDEX IF NOT EXISTS idx_batches_ingredient ON ingredient_batches(ingredient_id);
CREATE INDEX IF NOT EXISTS idx_batches_expiry ON ingredient_batches(expiry_date);
