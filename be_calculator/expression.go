package main

import (
	"encoding/json"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

const maxExpressionBytes = 1024
const maxExpressionDepth = 32

var decimalLiteral = regexp.MustCompile(`^(?:[0-9]+\.?[0-9]*|\.[0-9]+)(?:[eE][+-]?[0-9]+)?`)
var keypadSymbols = strings.NewReplacer("−", "-", "×", "*", "÷", "/")

func evaluate(w http.ResponseWriter, r *http.Request) {
	fields, ok := readOperationFields(w, r)
	if !ok {
		return
	}
	var input string
	if json.Unmarshal(fields["expression"], &input) != nil || len(input) > maxExpressionBytes {
		writeAPIError(w, http.StatusBadRequest, "INVALID_EXPRESSION", "Send an expression of at most 1024 bytes.")
		return
	}

	p := expressionParser{remaining: keypadSymbols.Replace(input)}
	result := p.sum()
	if p.invalid || p.peek() != 0 || p.remaining != "" {
		writeAPIError(w, http.StatusBadRequest, "INVALID_EXPRESSION", "Enter a valid expression using +, −, ×, or ÷.")
		return
	}
	writeOperationResult(w, result, p.calculationError)
}

// Each parsing level consumes operators of one precedence, from left to right.
type expressionParser struct {
	remaining        string
	depth            int
	invalid          bool
	calculationError *apiError
}

func (p *expressionParser) peek() byte {
	p.remaining = strings.TrimLeftFunc(p.remaining, unicode.IsSpace)
	if p.remaining == "" {
		return 0
	}
	return p.remaining[0]
}

func (p *expressionParser) sum() float64 {
	left := p.product()
	for op := p.peek(); !p.invalid && (op == '+' || op == '-'); op = p.peek() {
		p.remaining = p.remaining[1:]
		left = p.apply(op, left, p.product())
	}
	return left
}

func (p *expressionParser) product() float64 {
	left := p.operand()
	for op := p.peek(); !p.invalid && (op == '*' || op == '/'); op = p.peek() {
		p.remaining = p.remaining[1:]
		left = p.apply(op, left, p.operand())
	}
	return left
}

func (p *expressionParser) operand() float64 {
	sign := 1.0
	switch p.peek() {
	case '+':
		p.remaining = p.remaining[1:]
	case '-':
		sign = -1
		p.remaining = p.remaining[1:]
	}

	if p.peek() == '(' && p.depth < maxExpressionDepth {
		p.remaining = p.remaining[1:]
		p.depth++
		value := p.sum()
		p.depth--
		if p.peek() != ')' {
			p.invalid = true
			return 0
		}
		p.remaining = p.remaining[1:]
		return sign * value
	}

	literal := decimalLiteral.FindString(p.remaining)
	if literal == "" {
		p.invalid = true
		return 0
	}
	p.remaining = p.remaining[len(literal):]
	value, err := strconv.ParseFloat(literal, 64)
	if err != nil && p.calculationError == nil {
		p.calculationError = resultOutOfRange()
	}
	return sign * value
}

func (p *expressionParser) apply(op byte, left, right float64) float64 {
	if p.invalid || p.calculationError != nil {
		return 0
	}
	result, err := binaryOperations[op](left, right)
	if math.IsInf(result, 0) {
		err = resultOutOfRange()
	}
	p.calculationError = err
	return result
}
