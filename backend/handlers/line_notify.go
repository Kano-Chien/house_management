package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type LineNotifyHandler struct {
	DB *sql.DB
}

func (h *LineNotifyHandler) SendShoppingList(w http.ResponseWriter, r *http.Request) {
	// 1. Fetch shopping list data
	query := `
		WITH RequiredIngredients AS (
			SELECT
				ri.ingredient_id,
				SUM(ri.quantity) as total_required
			FROM recipe_ingredients ri
			JOIN meal_plan mp ON ri.recipe_id = mp.recipe_id
			GROUP BY ri.ingredient_id
		)
		SELECT
			i.name,
			(req.total_required - i.current_stock) as quantity_needed,
			COALESCE(i.unit, '') as unit,
			((req.total_required - i.current_stock) * COALESCE(i.price, 0)) as estimated_cost
		FROM ingredients i
		JOIN RequiredIngredients req ON i.id = req.ingredient_id
		WHERE i.current_stock < req.total_required
	`

	rows, err := h.DB.Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type ShoppingItem struct {
		Name           string  `json:"name"`
		QuantityNeeded float64 `json:"quantity_needed"`
		Unit           string  `json:"unit"`
		EstimatedCost  float64 `json:"estimated_cost"`
	}

	var items []ShoppingItem
	var totalCost float64
	for rows.Next() {
		var item ShoppingItem
		if err := rows.Scan(&item.Name, &item.QuantityNeeded, &item.Unit, &item.EstimatedCost); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		items = append(items, item)
		totalCost += item.EstimatedCost
	}

	// 2. Build message text
	if len(items) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "no_items", "message": "Everything is in stock!"})
		return
	}

	var sb strings.Builder
	sb.WriteString("🛒 Shopping List\n")
	sb.WriteString("━━━━━━━━━━━━━━━\n")
	for _, item := range items {
		line := fmt.Sprintf("• %s: %.0f %s", item.Name, item.QuantityNeeded, item.Unit)
		if item.EstimatedCost > 0 {
			line += fmt.Sprintf(" ($%.0f)", item.EstimatedCost)
		}
		sb.WriteString(line + "\n")
	}
	sb.WriteString("━━━━━━━━━━━━━━━\n")
	sb.WriteString(fmt.Sprintf("💰 Total: $%.0f", totalCost))

	// 3. Send via LINE Messaging API (Broadcast Message)
	token := os.Getenv("LINE_CHANNEL_ACCESS_TOKEN")
	// userID is not needed for broadcast
	// userID := os.Getenv("LINE_USER_ID")

	if token == "" {
		http.Error(w, "LINE credentials not configured", http.StatusInternalServerError)
		return
	}

	lineBody := map[string]interface{}{
		"messages": []map[string]string{
			{"type": "text", "text": sb.String()},
		},
	}

	bodyBytes, _ := json.Marshal(lineBody)
	// Use broadcast instead of push to avoid User ID issues
	req2, err := http.NewRequest("POST", "https://api.line.me/v2/bot/message/broadcast", bytes.NewBuffer(bodyBytes))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req2)
	if err != nil {
		http.Error(w, "Failed to send LINE message: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	requestID := resp.Header.Get("x-line-request-id")
	log.Printf("[LINE API] RequestID: %s | Status: %d | Body: %s", requestID, resp.StatusCode, string(respBody))

	if resp.StatusCode != 200 {
		http.Error(w, fmt.Sprintf("LINE API error (%d): %s", resp.StatusCode, string(respBody)), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "sent", "message": "Shopping list sent to LINE!"})
}
