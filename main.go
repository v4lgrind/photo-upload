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

var extByType = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
	"image/heic": ".heic",
	"image/heif": ".heic",
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func secHeaders(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; img-src 'self' data: blob:; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'")
		h.ServeHTTP(w, r)
	})
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
	log.Fatal(http.ListenAndServe(":"+port, secHeaders(mux)))
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

		buf := make([]byte, 512)
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			http.Error(w, "Failed to read file", http.StatusBadRequest)
			return
		}
		contentType := http.DetectContentType(buf[:n])
		if i := strings.Index(contentType, ";"); i >= 0 {
			contentType = strings.TrimSpace(contentType[:i])
		}
		ext, ok := extByType[contentType]
		if !ok {
			http.Error(w, "Only photos (JPEG, PNG, WEBP, GIF, HEIC) are allowed", http.StatusBadRequest)
			return
		}

		if seeker, ok := file.(io.Seeker); ok {
			seeker.Seek(0, io.SeekStart)
		}

		baseRaw := strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename))
		baseName := strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
				return r
			}
			return '_'
		}, baseRaw)
		if len(baseName) > 64 {
			baseName = baseName[:64]
		}
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
