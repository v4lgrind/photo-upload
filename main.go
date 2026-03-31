package main

import (
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

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	uploadDir := getEnv("UPLOAD_DIR", "/uploads")
	port := getEnv("PORT", "8080")
	maxSizeMB, _ := strconv.ParseInt(getEnv("MAX_FILE_SIZE", "300"), 10, 64)
	maxSize := maxSizeMB << 20

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatalf("Cannot create upload directory: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("static")))
	mux.HandleFunc("/upload", uploadHandler(uploadDir, maxSize))

	log.Printf("Listening on :%s, saving to %s (max %dMB per file)", port, uploadDir, maxSizeMB)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func uploadHandler(uploadDir string, maxSize int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxSize+1024)

		file, header, err := r.FormFile("photo")
		if err != nil {
			http.Error(w, "Failed to read file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Validate MIME type
		buf := make([]byte, 512)
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			http.Error(w, "Failed to read file", http.StatusBadRequest)
			return
		}
		contentType := http.DetectContentType(buf[:n])
		if !strings.HasPrefix(contentType, "image/") {
			http.Error(w, "Only image files are allowed", http.StatusBadRequest)
			return
		}

		// Seek back to start after reading for MIME detection
		if seeker, ok := file.(io.Seeker); ok {
			seeker.Seek(0, io.SeekStart)
		}

		// Generate unique filename
		ext := filepath.Ext(header.Filename)
		baseName := strings.TrimSuffix(header.Filename, ext)
		// Sanitize filename
		baseName = strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
				return r
			}
			return '_'
		}, baseName)
		filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), baseName, ext)
		destPath := filepath.Join(uploadDir, filename)

		dst, err := os.Create(destPath)
		if err != nil {
			http.Error(w, "Failed to save file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			os.Remove(destPath)
			http.Error(w, "Failed to save file", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":   "ok",
			"filename": filename,
		})
	}
}
