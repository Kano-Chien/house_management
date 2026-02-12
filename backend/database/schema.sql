-- Enable UUID extension if needed, or use SERIAL/INTEGER for simplicity in MVP
-- CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS ingredients (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    current_stock DECIMAL(10, 2) DEFAULT 0,
    unit VARCHAR(50),
    expiry_date DATE,
    price DECIMAL(10, 2) DEFAULT 0
);

-- Add price column if table already exists without it
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='ingredients' AND column_name='price') THEN
        ALTER TABLE ingredients ADD COLUMN price DECIMAL(10, 2) DEFAULT 0;
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS recipes (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    instructions TEXT
);

CREATE TABLE IF NOT EXISTS recipe_ingredients (
    recipe_id INTEGER REFERENCES recipes(id) ON DELETE CASCADE,
    ingredient_id INTEGER REFERENCES ingredients(id) ON DELETE CASCADE,
    quantity DECIMAL(10, 2) NOT NULL,
    PRIMARY KEY (recipe_id, ingredient_id)
);

CREATE TABLE IF NOT EXISTS meal_plan (
    id SERIAL PRIMARY KEY,
    date DATE NOT NULL,
    meal_type VARCHAR(50) NOT NULL CHECK (meal_type IN ('Lunch', 'Dinner')),
    recipe_id INTEGER REFERENCES recipes(id) ON DELETE SET NULL
);
