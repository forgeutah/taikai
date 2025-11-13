package api

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// UploadHandler handles file uploads
type UploadHandler struct {
	uploadDir string
	baseURL   string
}

// NewUploadHandler creates a new upload handler
func NewUploadHandler(uploadDir, baseURL string) *UploadHandler {
	return &UploadHandler{
		uploadDir: uploadDir,
		baseURL:   baseURL,
	}
}

// UploadResponse represents the upload response
type UploadResponse struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
}

// UploadImage handles POST /api/v1/upload/image
// Accepts multipart form data with "image" field
// Supports drag and drop functionality
func (h *UploadHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	// Limit upload size to 5MB
	r.Body = http.MaxBytesReader(w, r.Body, 5*1024*1024)

	err := r.ParseMultipartForm(5 * 1024 * 1024)
	if err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "File too large or invalid form data (max 5MB)")
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "No image file provided")
		return
	}
	defer file.Close()

	// Validate file type (images only)
	contentType := header.Header.Get("Content-Type")
	if !isValidImageType(contentType) {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid file type. Only JPEG, PNG, GIF, and WebP images are allowed")
		return
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = getExtensionFromContentType(contentType)
	}

	filename, err := generateUniqueFilename(ext)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to generate filename")
		return
	}

	// Ensure upload directory exists
	if err := os.MkdirAll(h.uploadDir, 0755); err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to create upload directory")
		return
	}

	// Create destination file
	destPath := filepath.Join(h.uploadDir, filename)
	dest, err := os.Create(destPath)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to create file")
		return
	}
	defer dest.Close()

	// Copy uploaded file to destination
	size, err := io.Copy(dest, file)
	if err != nil {
		os.Remove(destPath) // Clean up on error
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to save file")
		return
	}

	// Generate URL
	fileURL := fmt.Sprintf("%s/uploads/%s", h.baseURL, filename)

	response := UploadResponse{
		URL:      fileURL,
		Filename: filename,
		Size:     size,
	}

	RespondSuccess(w, http.StatusOK, response)
}

// isValidImageType checks if the content type is a valid image type
func isValidImageType(contentType string) bool {
	validTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"image/webp",
	}

	for _, validType := range validTypes {
		if strings.EqualFold(contentType, validType) {
			return true
		}
	}

	return false
}

// getExtensionFromContentType returns file extension based on content type
func getExtensionFromContentType(contentType string) string {
	switch strings.ToLower(contentType) {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		return ".jpg"
	}
}

// generateUniqueFilename generates a unique filename with the given extension
func generateUniqueFilename(ext string) (string, error) {
	// Generate 16 random bytes
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	// Convert to hex string
	filename := hex.EncodeToString(randomBytes) + ext

	return filename, nil
}
