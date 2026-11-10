package main

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	additionPath       = "/api/operations/addition"
	subtractionPath    = "/api/operations/subtraction"
	multiplicationPath = "/api/operations/multiplication"
	divisionPath       = "/api/operations/division"
	evaluatePath       = "/api/operations/evaluate"
)

func TestOperations(t *testing.T) {
	router := newRouter("")
	for _, tc := range []struct {
		name string
		path string
		body string
		want float64
	}{
		{"add to zero", additionPath, `{"left":0,"right":8}`, 8},
		{"add fractions and negatives", additionPath, `{"left":-2.25,"right":4.5}`, 2.25},
		{"add exponent with unknown field", additionPath, `{"left":1e3,"right":-1,"note":"ignored"}`, 999},
		{"subtract right from left", subtractionPath, `{"left":3,"right":8}`, -5},
		{"subtract a negative fraction", subtractionPath, `{"left":2.5,"right":-1.25}`, 3.75},
		{"multiply a negative fraction", multiplicationPath, `{"left":2.5,"right":-1.25}`, -3.125},
		{"multiply by zero", multiplicationPath, `{"left":7,"right":0}`, 0},
		{"divide left by right", divisionPath, `{"left":3,"right":12}`, 0.25},
		{"divide a negative fraction", divisionPath, `{"left":-3.75,"right":1.5}`, -2.5},
		{"divide zero", divisionPath, `{"left":0,"right":8}`, 0},
		{"mixed precedence", evaluatePath, `{"expression":"2 + 3 × 4 − 6 ÷ 2"}`, 11},
		{"left associative division", evaluatePath, `{"expression":"12 ÷ 3 ÷ 2"}`, 2},
		{"left associative subtraction", evaluatePath, `{"expression":"12-3-2"}`, 7},
		{"parentheses", evaluatePath, `{"expression":"(2+3)*4"}`, 20},
		{"signed decimals and scientific notation", evaluatePath, `{"expression":"-2.5*-2e+1 + .5"}`, 50.5},
		{"leading zero is decimal", evaluatePath, `{"expression":"08+1"}`, 9},
		{"fractional quotient", evaluatePath, `{"expression":"5/2"}`, 2.5},
		{"explicit percentage increase", evaluatePath, `{"expression":"200*(1+10%)"}`, 220},
		{"standalone percentage", evaluatePath, `{"expression":"50%"}`, 0.5},
		{"add a percentage", evaluatePath, `{"expression":"200+10%"}`, 200.1},
		{"subtract a percentage", evaluatePath, `{"expression":"200-10%"}`, 199.9},
		{"percentage of quantity", evaluatePath, `{"expression":"200*10%"}`, 20},
		{"divide by percentage", evaluatePath, `{"expression":"200/10%"}`, 2000},
		{"successive percentage additions", evaluatePath, `{"expression":"200+25%+25%"}`, 200.5},
		{"successive percentage subtractions", evaluatePath, `{"expression":"200-25%-25%"}`, 199.5},
		{"grouped percentage", evaluatePath, `{"expression":"200+(10%)"}`, 200.1},
		{"percentage group", evaluatePath, `{"expression":"200+(5+5)%"}`, 200.1},
		{"product is a numeric term", evaluatePath, `{"expression":"200+10%*2"}`, 200.2},
		{"grouped base", evaluatePath, `{"expression":"(100+100)+10%"}`, 200.1},
		{"negative percentage", evaluatePath, `{"expression":"200+-10%"}`, 199.9},
		{"percent in root", evaluatePath, `{"expression":"sqrt(25%)"}`, 0.5},
		{"percent after power", evaluatePath, `{"expression":"(2^3)%"}`, 0.08},
		{"power percent exponent", evaluatePath, `{"expression":"4^50%"}`, 2},
		{"zero percentage", evaluatePath, `{"expression":"0%"}`, 0},
		{"percentage multiplication", evaluatePath, `{"expression":"50%*20%"}`, 0.1},
		{"square root chain", evaluatePath, `{"expression":"2+sqrt(9+7)*3"}`, 14},
		{"nested square roots", evaluatePath, `{"expression":"sqrt(sqrt(81))^2"}`, 9},
		{"spaced square root", evaluatePath, `{"expression":"sqrt ( 0 )"}`, 0},
		{"square root endpoint", "/api/operations/square-root", `{"value":0.25}`, 0.5},
		{"mixed power chain", evaluatePath, `{"expression":"2+3^2*4"}`, 38},
		{"right associative powers", evaluatePath, `{"expression":"2^3^2"}`, 512},
		{"unary minus before power", evaluatePath, `{"expression":"-2^2"}`, -4},
		{"grouped negative base", evaluatePath, `{"expression":"(-2)^2"}`, 4},
		{"negative exponent", evaluatePath, `{"expression":"2^-2"}`, 0.25},
		{"fractional power", evaluatePath, `{"expression":"9^0.5"}`, 3},
		{"zero to zero", evaluatePath, `{"expression":"0^0"}`, 1},
		{"power endpoint", "/api/operations/exponentiation", `{"left":2,"right":3}`, 8},
		{"percentage endpoint", "/api/operations/percentage", `{"value":50}`, 0.5},
		{"unary plus on a group", evaluatePath, `{"expression":"2*+(3+4)"}`, 14},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tc.path, strings.NewReader(tc.body))
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != http.StatusOK || response.Header().Get("Content-Type") != "application/json" {
				t.Fatalf("expected 200 JSON, got status %d and headers %v", response.Code, response.Header())
			}
			var body map[string]float64
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if result, ok := body["result"]; !ok || math.Abs(result-tc.want) > 1e-12*math.Max(1, math.Abs(tc.want)) {
				t.Errorf("response = %v, want result %v", body, tc.want)
			}
		})
	}
}

func TestOperationErrors(t *testing.T) {
	router := newRouter("")
	for _, tc := range []struct {
		name   string
		path   string
		body   string
		status int
		code   string
	}{
		{"malformed JSON", additionPath, `{"left":1`, 400, "INVALID_JSON"},
		{"trailing JSON", additionPath, `{"left":1,"right":2} {}`, 400, "INVALID_JSON"},
		{"array", additionPath, `[]`, 400, "INVALID_JSON"},
		{"null body", additionPath, `null`, 400, "INVALID_JSON"},
		{"missing operand", additionPath, `{"left":1}`, 400, "INVALID_OPERANDS"},
		{"null operand", additionPath, `{"left":null,"right":2}`, 400, "INVALID_OPERANDS"},
		{"string operand", additionPath, `{"left":"1","right":2}`, 400, "INVALID_OPERANDS"},
		{"overflowing operand", additionPath, `{"left":1e309,"right":2}`, 400, "INVALID_OPERANDS"},
		{"overflowing sum", additionPath, `{"left":1e308,"right":1e308}`, 422, "RESULT_OUT_OF_RANGE"},
		{"oversized body", additionPath, strings.Repeat(" ", maxOperationBodyBytes+1), 413, "REQUEST_TOO_LARGE"},
		{"subtraction rejects invalid JSON", subtractionPath, `{"right":1`, 400, "INVALID_JSON"},
		{"subtraction rejects invalid operand", subtractionPath, `{"left":1,"right":"2"}`, 400, "INVALID_OPERANDS"},
		{"overflowing difference", subtractionPath, `{"left":1e308,"right":-1e308}`, 422, "RESULT_OUT_OF_RANGE"},
		{"subtraction rejects oversized body", subtractionPath, strings.Repeat(" ", maxOperationBodyBytes+1), 413, "REQUEST_TOO_LARGE"},
		{"multiplication rejects invalid operand", multiplicationPath, `{"left":"2","right":3}`, 400, "INVALID_OPERANDS"},
		{"overflowing product", multiplicationPath, `{"left":1e308,"right":2}`, 422, "RESULT_OUT_OF_RANGE"},
		{"validate before dividing", divisionPath, `{"left":null,"right":0}`, 400, "INVALID_OPERANDS"},
		{"divide by zero", divisionPath, `{"left":5,"right":0}`, 422, "DIVISION_BY_ZERO"},
		{"divide by negative zero", divisionPath, `{"left":5,"right":-0}`, 422, "DIVISION_BY_ZERO"},
		{"overflowing quotient", divisionPath, `{"left":1e308,"right":0.01}`, 422, "RESULT_OUT_OF_RANGE"},
		{"malformed expression body", evaluatePath, `{"expression":`, 400, "INVALID_JSON"},
		{"non-string expression", evaluatePath, `{"expression":2}`, 400, "INVALID_EXPRESSION"},
		{"incomplete expression", evaluatePath, `{"expression":"1/0+"}`, 400, "INVALID_EXPRESSION"},
		{"percentage missing operand", evaluatePath, `{"expression":"%"}`, 400, "INVALID_EXPRESSION"},
		{"repeated percentage", evaluatePath, `{"expression":"5%%"}`, 400, "INVALID_EXPRESSION"},
		{"unsupported operator", evaluatePath, `{"expression":"5!"}`, 400, "INVALID_EXPRESSION"},
		{"unsupported function call", evaluatePath, `{"expression":"sin(4)"}`, 400, "INVALID_EXPRESSION"},
		{"expression length limit", evaluatePath, `{"expression":"` + strings.Repeat("1", maxExpressionBytes+1) + `"}`, 400, "INVALID_EXPRESSION"},
		{"expression nesting limit", evaluatePath, `{"expression":"` + strings.Repeat("(", maxExpressionDepth+1) + "1" + strings.Repeat(")", maxExpressionDepth+1) + `"}`, 400, "INVALID_EXPRESSION"},
		{"divide by zero in group", evaluatePath, `{"expression":"2 + 3/(1-1)"}`, 422, "DIVISION_BY_ZERO"},
		{"intermediate overflow", evaluatePath, `{"expression":"1e308*2*0"}`, 422, "RESULT_OUT_OF_RANGE"},
		{"negative square root", evaluatePath, `{"expression":"sqrt(1-2)"}`, 422, "INVALID_DOMAIN"},
		{"square root syntax", evaluatePath, `{"expression":"sqrt9"}`, 400, "INVALID_EXPRESSION"},
		{"incomplete square root", evaluatePath, `{"expression":"sqrt("}`, 400, "INVALID_EXPRESSION"},
		{"square root empty", evaluatePath, `{"expression":"sqrt()"}`, 400, "INVALID_EXPRESSION"},
		{"square root propagates error", evaluatePath, `{"expression":"sqrt(1/0)"}`, 422, "DIVISION_BY_ZERO"},
		{"nested square root limit", evaluatePath, `{"expression":"` + strings.Repeat("sqrt(", maxExpressionDepth+1) + "1" + strings.Repeat(")", maxExpressionDepth+1) + `"}`, 400, "INVALID_EXPRESSION"},
		{"root malformed body", "/api/operations/square-root", `{`, 400, "INVALID_JSON"},
		{"root missing value", "/api/operations/square-root", `{}`, 400, "INVALID_OPERANDS"},
		{"root negative value", "/api/operations/square-root", `{"value":-1}`, 422, "INVALID_DOMAIN"},
		{"invalid power domain", evaluatePath, `{"expression":"(-2)^0.5"}`, 422, "INVALID_DOMAIN"},
		{"zero to negative power", evaluatePath, `{"expression":"0^-1"}`, 422, "DIVISION_BY_ZERO"},
		{"power overflow", evaluatePath, `{"expression":"10^400"}`, 422, "RESULT_OUT_OF_RANGE"},
		{"incomplete power", evaluatePath, `{"expression":"2^"}`, 400, "INVALID_EXPRESSION"},
		{"power recursion limit", evaluatePath, `{"expression":"` + strings.Repeat("1^", maxExpressionDepth+1) + `1"}`, 400, "INVALID_EXPRESSION"},
		{"overflowing literal", evaluatePath, `{"expression":"1e309+2"}`, 422, "RESULT_OUT_OF_RANGE"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tc.path, strings.NewReader(tc.body))
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
			if tc.code == "DIVISION_BY_ZERO" && body.Error["message"] != "Cannot divide by zero." {
				t.Errorf("message = %q, want Cannot divide by zero.", body.Error["message"])
			}
		})
	}
}
