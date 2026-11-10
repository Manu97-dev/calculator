package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestHealthRoute(t *testing.T) {
	router := newRouter("")
	for _, tc := range []struct {
		name   string
		method string
		path   string
		status int
	}{
		{"health check", http.MethodGet, "/api/health", http.StatusOK},
		{"unsupported method", http.MethodPost, "/api/health", http.StatusMethodNotAllowed},
		{"unknown path", http.MethodGet, "/api/missing", http.StatusNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(tc.method, tc.path, nil))
			if recorder.Code != tc.status {
				t.Fatalf("status = %d, want %d", recorder.Code, tc.status)
			}
			if tc.status != http.StatusOK {
				return
			}
			if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", contentType)
			}
			var body map[string]string
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if body["status"] != "ok" {
				t.Errorf("status field = %q, want ok", body["status"])
			}
		})
	}
}

func TestPackagedFrontend(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "assets"), 0755); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{
		"index.html":   "<!doctype html><title>Calculator</title>",
		"assets/ui.js": "console.log('calculator')",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}

	router := newRouter(dir)
	for _, tc := range []struct {
		method, path, accept string
		status               int
		body                 string
	}{
		{"GET", "/", "", 200, "<!doctype html><title>Calculator</title>"},
		{"GET", "/some/route", "text/html", 200, "<!doctype html><title>Calculator</title>"},
		{"GET", "/assets/ui.js", "", 200, "console.log('calculator')"},
		{"GET", "/assets/missing.js", "text/html", 404, ""},
		{"GET", "/some/route", "application/json", 404, ""},
		{"GET", "/api/health", "text/html", 200, "{\"status\":\"ok\"}\n"},
		{"POST", "/api/health", "", 405, ""},
		{"GET", "/api", "text/html", 404, ""},
		{"GET", "/api/missing", "text/html", 404, ""},
		{"GET", additionPath, "text/html", 405, ""},
		{"GET", subtractionPath, "text/html", 405, ""},
	} {
		t.Run(tc.method+" "+tc.path+" "+tc.accept, func(t *testing.T) {
			request := httptest.NewRequest(tc.method, tc.path, nil)
			request.Header.Set("Accept", tc.accept)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != tc.status {
				t.Fatalf("status = %d, want %d", response.Code, tc.status)
			}
			if tc.body != "" && response.Body.String() != tc.body {
				t.Errorf("body = %q, want %q", response.Body.String(), tc.body)
			}
		})
	}
}
