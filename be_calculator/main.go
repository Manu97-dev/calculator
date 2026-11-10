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
	api := http.NewServeMux()
	api.HandleFunc("GET /api/health", health)
	registerOperations(api)

	mux := http.NewServeMux()
	mux.HandleFunc("/api", http.NotFound)
	mux.Handle("/api/", api)
	if staticDir != "" {
		spa := http.NewServeMux()
		spa.Handle("GET /", frontend(staticDir))
		mux.Handle("/", spa)
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
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("write JSON response: %v", err)
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
