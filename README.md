# House Management 🏠

A full-stack home management web application designed to run on a **Raspberry Pi**. It helps you manage your household inventory, plan weekly meals, track recipes, and generate shopping lists.

## Features

- 📦 **Inventory Management** — Track your pantry stock levels
- 🍽️ **Meal Planner** — Schedule meals for the week
- 📖 **Recipe Manager** — Store recipes with ingredients
- 🛒 **Shopping List** — Auto-generated from meal plan, with LINE Notify integration
- 📷 **Image Upload** — Attach photos to inventory items

## Tech Stack

| Layer    | Technology |
|----------|------------|
| Frontend | Vue 3 + Vite + TailwindCSS |
| Backend  | Go (net/http) |
| Database | SQLite |
| Deploy   | Raspberry Pi + systemd |

## Project Structure

```
house_management/
├── backend/
│   ├── main.go              # Entry point, router setup
│   ├── handlers/            # HTTP handlers (inventory, recipes, meal plan, etc.)
│   ├── models/              # Data models
│   ├── database/            # SQLite schema
│   └── uploads/             # Uploaded images
├── frontend/
│   ├── src/
│   │   └── components/      # Vue components
│   └── dist/                # Built frontend (served by Go)
└── deploy/
    ├── systemd/             # systemd service files
    │   └── house-backend.service
    ├── scripts/
    │   └── monitor_backend.sh  # Backend health monitor daemon
    └── README.md            # Raspberry Pi deployment guide
```

## Development Setup

### Prerequisites
- Go 1.21+
- Node.js 18+

### Backend
```bash
cd backend
go run main.go
```

### Frontend
```bash
cd frontend
npm install
npm run dev
```

The frontend dev server proxies `/api` requests to the Go backend at `http://localhost:8080`.

## Building for Production

### 1. Build the Frontend
```bash
cd frontend
npm run build
```
The built files will be placed in `frontend/dist/`, which the Go backend serves automatically.

### 2. Build the Go Backend

For **local machine**:
```bash
go build -o backend/house-backend ./backend
```

For **Raspberry Pi (arm64)**:
```bash
GOOS=linux GOARCH=arm64 go build -o backend/backend_pi ./backend
```

For **Raspberry Pi (arm, older 32-bit OS)**:
```bash
GOOS=linux GOARCH=arm go build -o backend/backend_pi ./backend
```

## Deploying to Raspberry Pi

See [deploy/README.md](deploy/README.md) for complete step-by-step deployment instructions, including:
- Copying files to the Pi
- Setting up the `systemd` service to auto-start on boot
- Running the optional health monitor daemon

## Environment Variables

Create a `.env` file in the `backend/` directory:
```env
DATABASE_URL=house.db
LINE_NOTIFY_TOKEN=your_line_notify_token_here
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET/POST | `/api/inventory` | List / add ingredients |
| PUT | `/api/inventory/stock` | Update stock level |
| PUT | `/api/inventory/edit` | Edit ingredient |
| POST | `/api/inventory/delete` | Delete ingredient |
| GET/POST | `/api/recipes` | List / create recipes |
| GET/POST | `/api/mealplan` | Get / schedule meal plan |
| POST | `/api/mealplan/cook` | Mark meal as cooked |
| GET/POST | `/api/shopping-list` | Get / add items to shopping list |
| POST | `/api/line/send-shopping-list` | Send shopping list via LINE Notify |
| POST | `/api/upload` | Upload image |
