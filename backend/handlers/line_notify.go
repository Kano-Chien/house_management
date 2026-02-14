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
	// 1. Parse request body (list of items from frontend)
	type RequestItem struct {
		Name string `json:"name"`
	}
	var items []RequestItem

	if err := json.NewDecoder(r.Body).Decode(&items); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(items) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "no_items", "message": "List is empty!"})
		return
	}

	// 2. Build message text
	var sb strings.Builder
	sb.WriteString("🛒 Shopping List\n")
	sb.WriteString("━━━━━━━━━━━━━━━\n")
	for _, item := range items {
		sb.WriteString(fmt.Sprintf("• %s\n", item.Name))
	}

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
