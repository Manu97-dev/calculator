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

func registerOperations(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/operations/addition", operationHandler(func(left, right float64) (float64, *apiError) {
		return left + right, nil
	}))
	mux.HandleFunc("POST /api/operations/subtraction", operationHandler(func(left, right float64) (float64, *apiError) {
		return left - right, nil
	}))
	mux.HandleFunc("POST /api/operations/multiplication", operationHandler(func(left, right float64) (float64, *apiError) {
		return left * right, nil
	}))
	mux.HandleFunc("POST /api/operations/division", operationHandler(func(left, right float64) (float64, *apiError) {
		if right == 0 {
			return 0, &apiError{Code: "DIVISION_BY_ZERO", Message: "Cannot divide by zero."}
		}
		return left / right, nil
	}))
}

func operationHandler(calculate func(left, right float64) (float64, *apiError)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxOperationBodyBytes))
		var sizeError *http.MaxBytesError
		if errors.As(err, &sizeError) {
			writeAPIError(w, http.StatusRequestEntityTooLarge, "REQUEST_TOO_LARGE", "The request body must be 4 KiB or smaller.")
			return
		}
		var fields map[string]json.RawMessage
		if err != nil || json.Unmarshal(body, &fields) != nil || fields == nil {
			writeAPIError(w, http.StatusBadRequest, "INVALID_JSON", "Send a single JSON object.")
			return
		}

		left, leftOK := parseOperand(fields["left"])
		right, rightOK := parseOperand(fields["right"])
		if !leftOK || !rightOK {
			writeAPIError(w, http.StatusBadRequest, "INVALID_OPERANDS", "Both left and right must be finite JSON numbers.")
			return
		}

		result, calculationError := calculate(left, right)
		if calculationError != nil {
			writeAPIError(w, http.StatusUnprocessableEntity, calculationError.Code, calculationError.Message)
			return
		}
		if math.IsInf(result, 0) {
			writeAPIError(w, http.StatusUnprocessableEntity, "RESULT_OUT_OF_RANGE", "The result is outside the supported numeric range.")
			return
		}
		writeJSON(w, http.StatusOK, operationResponse{Result: result})
	}
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
