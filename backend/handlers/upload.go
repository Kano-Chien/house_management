package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// UploadImage handles image uploads for the recipe rich text editor
func UploadImage(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form with a max memory of 10MB
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Retrieve the file from form data
	file, handler, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Error retrieving image file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate MIME type — only allow safe image formats
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	contentType := handler.Header.Get("Content-Type")
	if !allowedTypes[contentType] {
		http.Error(w, "Only JPEG, PNG, GIF, and WebP images are allowed", http.StatusBadRequest)
		return
	}

	// Ensure uploads directory exists
	uploadDir := "./uploads"
	err = os.MkdirAll(uploadDir, os.ModePerm)
	if err != nil {
		log.Printf("Could not create upload directory: %v", err)
		http.Error(w, "Could not create upload directory", http.StatusInternalServerError)
		return
	}

	// Create a unique filename with safe base name to prevent path traversal
	baseName := filepath.Base(handler.Filename)
	safeName := time.Now().Format("20060102150405") + "_" + strings.ReplaceAll(baseName, " ", "_")
	dstPath := filepath.Join(uploadDir, safeName)

	dst, err := os.Create(dstPath)
	if err != nil {
		log.Printf("Could not save file: %v", err)
		http.Error(w, "Could not save file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		log.Printf("Could not save file content: %v", err)
		http.Error(w, "Could not save file content", http.StatusInternalServerError)
		return
	}

	// Return the relative URL to access the image
	imageURL := "/uploads/" + safeName

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"url": imageURL,
	})
}
