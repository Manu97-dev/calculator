package main

import (
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type healthResponse struct {
	Status string `json:"status"`
}

func newRouter(staticDir string) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", health)
	if staticDir != "" {
		mux.HandleFunc("GET /api", http.NotFound)
		mux.HandleFunc("GET /api/", http.NotFound)
		mux.Handle("GET /", frontend(staticDir))
	}
	return mux
}

func frontend(dir string) http.Handler {
	files := os.DirFS(dir)
	fileServer := http.FileServerFS(files)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/")
		if name != "" {
			if info, err := fs.Stat(files, name); err == nil && !info.IsDir() {
				fileServer.ServeHTTP(w, r)
				return
			}
			if strings.HasPrefix(name, "assets/") || !strings.Contains(r.Header.Get("Accept"), "text/html") {
				http.NotFound(w, r)
				return
			}
		}
		http.ServeFile(w, r, filepath.Join(dir, "index.html"))
	})
}

func health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(healthResponse{Status: "ok"}); err != nil {
		log.Printf("write health response: %v", err)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	staticDir := os.Getenv("STATIC_DIR")
	if staticDir != "" {
		index, err := os.Stat(filepath.Join(staticDir, "index.html"))
		if err != nil || index.IsDir() {
			log.Fatalf("STATIC_DIR must contain index.html: %s", staticDir)
		}
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           newRouter(staticDir),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("Listening on port %s", port)
	log.Fatal(server.ListenAndServe())
}
