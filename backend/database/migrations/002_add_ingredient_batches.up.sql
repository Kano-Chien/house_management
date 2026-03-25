CREATE TABLE IF NOT EXISTS ingredient_batches (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ingredient_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    expiry_date TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (ingredient_id) REFERENCES ingredients(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_batches_ingredient ON ingredient_batches(ingredient_id);
CREATE INDEX IF NOT EXISTS idx_batches_expiry ON ingredient_batches(expiry_date);
