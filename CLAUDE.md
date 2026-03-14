# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A full-stack home management app (inventory, recipes, meal planning, shopping list) designed to run on a Raspberry Pi. Go backend + Vue 3 frontend + SQLite.

## Development Commands

### Backend
```bash
cd backend
go run main.go          # Run dev server on :8080
go build -o house_backend main.go  # Build binary
```

### Frontend
```bash
cd frontend
npm install
npm run dev             # Dev server on :5173, proxies /api and /uploads to :8080
npm run build           # Build to frontend/dist/ (served by Go in production)
```

### Deploy to Raspberry Pi
```bash
./deploy/scripts/deploy.sh          # Builds both, deploys via rsync, restarts systemd service
./deploy/scripts/deploy.sh --force-env  # Also re-deploy .env file
```

### Cross-compile for Raspberry Pi
```bash
# 64-bit ARM
GOOS=linux GOARCH=arm64 go build -o backend_pi main.go

# 32-bit ARM (older Pi)
GOOS=linux GOARCH=arm GOARM=6 CGO_ENABLED=0 go build -o backend_pi main.go
```

## Architecture

### Backend (Go, `backend/`)
- **No framework** — uses stdlib `net/http` with `http.NewServeMux()`
- `main.go` sets up routes, CORS middleware, static file serving, and DB connection
- `handlers/` — one file per domain (inventory, recipe, meal_plan, shopping_list, line_notify, upload)
- `models/` — plain structs, no ORM
- `database/schema_sqlite.sql` — schema applied on startup via embed
- Uses `modernc.org/sqlite` (pure Go, no CGO) — critical for cross-compilation to ARM
- SQLite connection pool set to `SetMaxOpenConns(1)` to prevent lock contention on Pi
- WAL mode and foreign keys enabled via PRAGMAs

### Frontend (Vue 3 + Vite, `frontend/`)
- Single-page app with tab-based navigation: Inventory | Recipes | Meal Plan | Shopping List
- Active tab persisted to `localStorage`
- Mobile-first: bottom nav on mobile, horizontal nav on desktop
- `App.vue` — root component, tab switching, mobile/desktop layout
- Components communicate with backend via `fetch()` to `/api/*`
- Tiptap used for rich text editing in recipe notes/steps
- TailwindCSS for styling

### Database Schema (SQLite)
- `ingredients` — `name, current_stock, min_stock, price, category, is_tracked`
- `recipes` — `name, notes`
- `recipe_ingredients` — junction table `(recipe_id, ingredient_id, quantity)` with CASCADE delete
- `meal_plan` — `date, meal_type, recipe_id (nullable), custom_name, is_cooked`
- `shopping_list_items` — `name, ingredient_id (nullable FK), is_custom, is_checked`

**Key business logic:** Shopping list auto-syncs — ingredients below `min_stock` are automatically added; restored when stock is replenished. This happens in the `shopping_list.go` handler on stock updates.

### Configuration
Backend reads from `backend/.env` (see `backend/.env.example`):
- `DATABASE_URL` — SQLite file path (default: `house.db`)
- `LINE_CHANNEL_ACCESS_TOKEN` / `LINE_USER_ID` — LINE Notify integration
- `CORS_ALLOWED_ORIGINS` — comma-separated origins (defaults to `*`)

### External Integration
- **LINE Notify**: Sends shopping lists to LINE app via LINE Messaging API (`handlers/line_notify.go`)

### Production Deployment
- Go binary serves both API (`/api/*`) and built frontend (`frontend/dist/`) as static files
- SPA fallback: all unmatched routes serve `index.html`
- Runs as systemd service (`deploy/systemd/house-backend.service`) with auto-restart
