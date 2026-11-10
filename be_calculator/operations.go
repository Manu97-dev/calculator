package main

import (
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
)

const maxOperationBodyBytes = 4 * 1024

type operationResponse struct {
	Result float64 `json:"result"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

var binaryOperations = map[byte]func(left, right float64) (float64, *apiError){
	'+': func(left, right float64) (float64, *apiError) { return left + right, nil },
	'-': func(left, right float64) (float64, *apiError) { return left - right, nil },
	'*': func(left, right float64) (float64, *apiError) { return left * right, nil },
	'^': func(left, right float64) (float64, *apiError) {
		if left == 0 && right < 0 {
			return 0, &apiError{Code: "DIVISION_BY_ZERO", Message: "Cannot divide by zero."}
		}
		result := math.Pow(left, right)
		if math.IsNaN(result) {
			return 0, &apiError{Code: "INVALID_DOMAIN", Message: "This power has no real-number result."}
		}
		return result, nil
	},
	'/': func(left, right float64) (float64, *apiError) {
		if right == 0 {
			return 0, &apiError{Code: "DIVISION_BY_ZERO", Message: "Cannot divide by zero."}
		}
		return left / right, nil
	},
}

func registerOperations(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/operations/addition", operationHandler(binaryOperations['+']))
	mux.HandleFunc("POST /api/operations/subtraction", operationHandler(binaryOperations['-']))
	mux.HandleFunc("POST /api/operations/multiplication", operationHandler(binaryOperations['*']))
	mux.HandleFunc("POST /api/operations/division", operationHandler(binaryOperations['/']))
	mux.HandleFunc("POST /api/operations/exponentiation", operationHandler(binaryOperations['^']))
	mux.HandleFunc("POST /api/operations/square-root", squareRootHandler)
	mux.HandleFunc("POST /api/operations/evaluate", evaluate)
}

func squareRoot(value float64) (float64, *apiError) {
	if value < 0 {
		return 0, &apiError{Code: "INVALID_DOMAIN", Message: "Square root requires a nonnegative number."}
	}
	return math.Sqrt(value), nil
}

func squareRootHandler(w http.ResponseWriter, r *http.Request) {
	fields, ok := readOperationFields(w, r)
	if !ok {
		return
	}
	value, valid := parseOperand(fields["value"])
	if !valid {
		writeAPIError(w, http.StatusBadRequest, "INVALID_OPERANDS", "Value must be a finite JSON number.")
		return
	}
	result, calculationError := squareRoot(value)
	writeOperationResult(w, result, calculationError)
}

func operationHandler(calculate func(left, right float64) (float64, *apiError)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fields, ok := readOperationFields(w, r)
		if !ok {
			return
		}
		left, leftOK := parseOperand(fields["left"])
		right, rightOK := parseOperand(fields["right"])
		if !leftOK || !rightOK {
			writeAPIError(w, http.StatusBadRequest, "INVALID_OPERANDS", "Both left and right must be finite JSON numbers.")
			return
		}

		result, calculationError := calculate(left, right)
		writeOperationResult(w, result, calculationError)
	}
}

func readOperationFields(w http.ResponseWriter, r *http.Request) (map[string]json.RawMessage, bool) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxOperationBodyBytes))
	var sizeError *http.MaxBytesError
	if errors.As(err, &sizeError) {
		writeAPIError(w, http.StatusRequestEntityTooLarge, "REQUEST_TOO_LARGE", "The request body must be 4 KiB or smaller.")
		return nil, false
	}
	var fields map[string]json.RawMessage
	if err != nil || json.Unmarshal(body, &fields) != nil || fields == nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_JSON", "Send a single JSON object.")
		return nil, false
	}
	return fields, true
}

func resultOutOfRange() *apiError {
	return &apiError{Code: "RESULT_OUT_OF_RANGE", Message: "The result is outside the supported numeric range."}
}

func writeOperationResult(w http.ResponseWriter, result float64, calculationError *apiError) {
	if calculationError == nil && math.IsInf(result, 0) {
		calculationError = resultOutOfRange()
	}
	if calculationError != nil {
		writeAPIError(w, http.StatusUnprocessableEntity, calculationError.Code, calculationError.Message)
		return
	}
	writeJSON(w, http.StatusOK, operationResponse{Result: result})
}

func parseOperand(raw json.RawMessage) (float64, bool) {
	var value *float64
	if err := json.Unmarshal(raw, &value); err != nil || value == nil {
		return 0, false
	}
	return *value, true
}

func writeAPIError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, struct {
		Error apiError `json:"error"`
	}{Error: apiError{Code: code, Message: message}})
}
