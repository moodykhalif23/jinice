package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	maxUploadSize = 10 << 20 // 10 MB
	uploadDir     = "./uploads"
)

type Image struct {
	ID            int       `json:"id"`
	EntityType    string    `json:"entity_type"`
	EntityID      int       `json:"entity_id"`
	ImageURL      string    `json:"image_url"`
	StoragePath   string    `json:"storage_path,omitempty"`
	Caption       string    `json:"caption,omitempty"`
	DisplayOrder  int       `json:"display_order"`
	IsPrimary     bool      `json:"is_primary"`
	UploadedBy    *int      `json:"uploaded_by,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type ImageMetadata struct {
	ID               int    `json:"id"`
	ImageID          int    `json:"image_id"`
	FileSize         int    `json:"file_size,omitempty"`
	Width            int    `json:"width,omitempty"`
	Height           int    `json:"height,omitempty"`
	MimeType         string `json:"mime_type,omitempty"`
	OriginalFilename string `json:"original_filename,omitempty"`
}

// Initialize upload directory
func InitImageStorage() error {
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return fmt.Errorf("failed to create upload directory: %v", err)
	}
	return nil
}

// Get images for an entity (business or event)
func getImagesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	entityType := r.URL.Query().Get("entity_type")
	entityIDStr := r.URL.Query().Get("entity_id")

	if entityType == "" || entityIDStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "entity_type and entity_id are required"})
		return
	}

	entityID, err := strconv.Atoi(entityIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid entity_id"})
		return
	}

	rows, err := db.Query(`
		SELECT id, entity_type, entity_id, image_url, storage_path, caption, display_order, is_primary, uploaded_by, created_at
		FROM images
		WHERE entity_type = ? AND entity_id = ?
		ORDER BY is_primary DESC, display_order ASC, created_at ASC
	`, entityType, entityID)

	if err != nil {
		log.Printf("Error querying images: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
		return
	}
	defer rows.Close()

	var images []Image
	for rows.Next() {
		var img Image
		var storagePath, caption sql.NullString
		var uploadedBy sql.NullInt64

		err := rows.Scan(&img.ID, &img.EntityType, &img.EntityID, &img.ImageURL, &storagePath, &caption, &img.DisplayOrder, &img.IsPrimary, &uploadedBy, &img.CreatedAt)
		if err != nil {
			log.Printf("Error scanning image: %v", err)
			continue
		}

		if storagePath.Valid {
			img.StoragePath = storagePath.String
		}
		if caption.Valid {
			img.Caption = caption.String
		}
		if uploadedBy.Valid {
			uid := int(uploadedBy.Int64)
			img.UploadedBy = &uid
		}

		images = append(images, img)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(images)
}

// Upload image for an entity
func uploadImageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "file too large or invalid form"})
		return
	}

	// Get form values
	entityType := r.FormValue("entity_type")
	entityIDStr := r.FormValue("entity_id")
	caption := r.FormValue("caption")
	isPrimaryStr := r.FormValue("is_primary")

	if entityType == "" || entityIDStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "entity_type and entity_id are required"})
		return
	}

	entityID, err := strconv.Atoi(entityIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid entity_id"})
		return
	}

	isPrimary := isPrimaryStr == "true"

	// Get user ID from auth
	userIDStr := r.Header.Get("X-User-ID")
	var uploadedBy *int
	if userIDStr != "" {
		uid, err := strconv.Atoi(userIDStr)
		if err == nil {
			uploadedBy = &uid
		}
	}

	// Get file from form
	file, header, err := r.FormFile("image")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "image file is required"})
		return
	}
	defer file.Close()

	// Validate file type
	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "file must be an image"})
		return
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%s_%d_%d%s", entityType, entityID, time.Now().Unix(), ext)
	filepath := filepath.Join(uploadDir, filename)

	// Create file
	dst, err := os.Create(filepath)
	if err != nil {
		log.Printf("Error creating file: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to save file"})
		return
	}
	defer dst.Close()

	// Copy file content
	fileSize, err := io.Copy(dst, file)
	if err != nil {
		log.Printf("Error copying file: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to save file"})
		return
	}

	// Generate URL for the uploaded file
	imageURL := fmt.Sprintf("/uploads/%s", filename)

	// If this is primary, unset other primary images
	if isPrimary {
		_, err = db.Exec("UPDATE images SET is_primary = FALSE WHERE entity_type = ? AND entity_id = ?", entityType, entityID)
		if err != nil {
			log.Printf("Error unsetting primary images: %v", err)
		}
	}

	// Get next display order
	var maxOrder sql.NullInt64
	err = db.QueryRow("SELECT MAX(display_order) FROM images WHERE entity_type = ? AND entity_id = ?", entityType, entityID).Scan(&maxOrder)
	displayOrder := 0
	if maxOrder.Valid {
		displayOrder = int(maxOrder.Int64) + 1
	}

	// Insert image record
	result, err := db.Exec(`
		INSERT INTO images (entity_type, entity_id, image_url, storage_path, caption, display_order, is_primary, uploaded_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, entityType, entityID, imageURL, filepath, caption, displayOrder, isPrimary, uploadedBy)

	if err != nil {
		log.Printf("Error inserting image record: %v", err)
		os.Remove(filepath) // Clean up file
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to save image record"})
		return
	}

	imageID, _ := result.LastInsertId()

	// Insert metadata
	_, err = db.Exec(`
		INSERT INTO image_metadata (image_id, file_size, mime_type, original_filename)
		VALUES (?, ?, ?, ?)
	`, imageID, fileSize, contentType, header.Filename)

	if err != nil {
		log.Printf("Error inserting image metadata: %v", err)
	}

	// Return created image
	image := Image{
		ID:           int(imageID),
		EntityType:   entityType,
		EntityID:     entityID,
		ImageURL:     imageURL,
		StoragePath:  filepath,
		Caption:      caption,
		DisplayOrder: displayOrder,
		IsPrimary:    isPrimary,
		UploadedBy:   uploadedBy,
		CreatedAt:    time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(image)
}

// Add image by URL (for external images)
func addImageURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		EntityType string `json:"entity_type"`
		EntityID   int    `json:"entity_id"`
		ImageURL   string `json:"image_url"`
		Caption    string `json:"caption"`
		IsPrimary  bool   `json:"is_primary"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}

	if req.EntityType == "" || req.EntityID == 0 || req.ImageURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "entity_type, entity_id, and image_url are required"})
		return
	}

	// Get user ID from auth
	userIDStr := r.Header.Get("X-User-ID")
	var uploadedBy *int
	if userIDStr != "" {
		uid, err := strconv.Atoi(userIDStr)
		if err == nil {
			uploadedBy = &uid
		}
	}

	// If this is primary, unset other primary images
	if req.IsPrimary {
		_, err := db.Exec("UPDATE images SET is_primary = FALSE WHERE entity_type = ? AND entity_id = ?", req.EntityType, req.EntityID)
		if err != nil {
			log.Printf("Error unsetting primary images: %v", err)
		}
	}

	// Get next display order
	var maxOrder sql.NullInt64
	err := db.QueryRow("SELECT MAX(display_order) FROM images WHERE entity_type = ? AND entity_id = ?", req.EntityType, req.EntityID).Scan(&maxOrder)
	displayOrder := 0
	if maxOrder.Valid {
		displayOrder = int(maxOrder.Int64) + 1
	}

	// Insert image record
	result, err := db.Exec(`
		INSERT INTO images (entity_type, entity_id, image_url, caption, display_order, is_primary, uploaded_by)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, req.EntityType, req.EntityID, req.ImageURL, req.Caption, displayOrder, req.IsPrimary, uploadedBy)

	if err != nil {
		log.Printf("Error inserting image record: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to save image record"})
		return
	}

	imageID, _ := result.LastInsertId()

	image := Image{
		ID:           int(imageID),
		EntityType:   req.EntityType,
		EntityID:     req.EntityID,
		ImageURL:     req.ImageURL,
		Caption:      req.Caption,
		DisplayOrder: displayOrder,
		IsPrimary:    req.IsPrimary,
		UploadedBy:   uploadedBy,
		CreatedAt:    time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(image)
}

// Update image
func updateImageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID           int    `json:"id"`
		Caption      string `json:"caption"`
		DisplayOrder int    `json:"display_order"`
		IsPrimary    bool   `json:"is_primary"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}

	// Get image to check entity info
	var entityType string
	var entityID int
	err := db.QueryRow("SELECT entity_type, entity_id FROM images WHERE id = ?", req.ID).Scan(&entityType, &entityID)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "image not found"})
			return
		}
		log.Printf("Error fetching image: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
		return
	}

	// If setting as primary, unset other primary images
	if req.IsPrimary {
		_, err = db.Exec("UPDATE images SET is_primary = FALSE WHERE entity_type = ? AND entity_id = ? AND id != ?", entityType, entityID, req.ID)
		if err != nil {
			log.Printf("Error unsetting primary images: %v", err)
		}
	}

	// Update image
	_, err = db.Exec(`
		UPDATE images 
		SET caption = ?, display_order = ?, is_primary = ?
		WHERE id = ?
	`, req.Caption, req.DisplayOrder, req.IsPrimary, req.ID)

	if err != nil {
		log.Printf("Error updating image: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to update image"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "image updated successfully"})
}

// Delete image
func deleteImageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID int `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}

	// Get image info before deletion
	var storagePath sql.NullString
	err := db.QueryRow("SELECT storage_path FROM images WHERE id = ?", req.ID).Scan(&storagePath)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "image not found"})
			return
		}
		log.Printf("Error fetching image: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
		return
	}

	// Delete from database
	_, err = db.Exec("DELETE FROM images WHERE id = ?", req.ID)
	if err != nil {
		log.Printf("Error deleting image: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to delete image"})
		return
	}

	// Delete file if it exists locally
	if storagePath.Valid && storagePath.String != "" {
		if err := os.Remove(storagePath.String); err != nil {
			log.Printf("Warning: Could not delete file %s: %v", storagePath.String, err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "image deleted successfully"})
}
