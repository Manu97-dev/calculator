package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const additionPath = "/api/operations/addition"

func TestAddition(t *testing.T) {
	router := newRouter("")
	for _, tc := range []struct {
		name string
		body string
		want float64
	}{
		{"zero", `{"left":0,"right":8}`, 8},
		{"fractions and negatives", `{"left":-2.25,"right":4.5}`, 2.25},
		{"exponent and unknown field", `{"left":1e3,"right":-1,"note":"ignored"}`, 999},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, additionPath, strings.NewReader(tc.body))
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != http.StatusOK || response.Header().Get("Content-Type") != "application/json" {
				t.Fatalf("expected 200 JSON, got status %d and headers %v", response.Code, response.Header())
			}
			var body map[string]float64
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if result, ok := body["result"]; !ok || result != tc.want {
				t.Errorf("response = %v, want result %v", body, tc.want)
			}
		})
	}
}

func TestAdditionErrors(t *testing.T) {
	router := newRouter("")
	for _, tc := range []struct {
		name   string
		body   string
		status int
		code   string
	}{
		{"malformed JSON", `{"left":1`, 400, "INVALID_JSON"},
		{"trailing JSON", `{"left":1,"right":2} {}`, 400, "INVALID_JSON"},
		{"array", `[]`, 400, "INVALID_JSON"},
		{"null body", `null`, 400, "INVALID_JSON"},
		{"missing operand", `{"left":1}`, 400, "INVALID_OPERANDS"},
		{"null operand", `{"left":null,"right":2}`, 400, "INVALID_OPERANDS"},
		{"string operand", `{"left":"1","right":2}`, 400, "INVALID_OPERANDS"},
		{"overflowing operand", `{"left":1e309,"right":2}`, 400, "INVALID_OPERANDS"},
		{"overflowing sum", `{"left":1e308,"right":1e308}`, 422, "RESULT_OUT_OF_RANGE"},
		{"oversized body", strings.Repeat(" ", maxOperationBodyBytes+1), 413, "REQUEST_TOO_LARGE"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, additionPath, strings.NewReader(tc.body))
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != tc.status || response.Header().Get("Content-Type") != "application/json" {
				t.Fatalf("expected %d JSON, got status %d and headers %v", tc.status, response.Code, response.Header())
			}
			var body struct {
				Error map[string]string `json:"error"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Error["code"] != tc.code || body.Error["message"] == "" {
				t.Errorf("error = %v, want code %q and a message", body.Error, tc.code)
			}
		})
	}
}
