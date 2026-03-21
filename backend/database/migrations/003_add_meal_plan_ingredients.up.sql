CREATE TABLE IF NOT EXISTS meal_plan_ingredients (
    meal_plan_id INTEGER,
    ingredient_id INTEGER,
    quantity INTEGER NOT NULL,
    PRIMARY KEY (meal_plan_id, ingredient_id),
    FOREIGN KEY (meal_plan_id) REFERENCES meal_plan(id) ON DELETE CASCADE,
    FOREIGN KEY (ingredient_id) REFERENCES ingredients(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_meal_plan_ingredients ON meal_plan_ingredients(meal_plan_id);
