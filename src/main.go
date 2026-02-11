package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type ProductDetails struct {
	ProductID     int32  `json:"product_id"`
	SKU           string `json:"sku"`
	Manufacturer  string `json:"manufacturer"`
	CategoryID    int32  `json:"category_id"`
	Weight        int32  `json:"weight"`
	SomeOtherID   int32  `json:"some_other_id"`
}

type ApiError struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Details string `json:"details"`
}

var (
	mu       sync.RWMutex
	products = map[int32]ProductDetails{}
)

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, errCode, msg, details string) {
	writeJSON(w, code, ApiError{Error: errCode, Message: msg, Details: details})
}

// GET /products/{productId}
// POST /products/{productId}/details
func handler(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	// Expect: products/{id} or products/{id}/details
	if len(parts) < 2 || parts[0] != "products" {
		writeErr(w, http.StatusNotFound, "NOT_FOUND", "Route not found", "Unknown path")
		return
	}

	// parse productId
	id64, err := strconv.ParseInt(parts[1], 10, 32)
	if err != nil || id64 <= 0 {
		writeErr(w, http.StatusBadRequest, "INVALID_INPUT", "The provided input data is invalid", "Product ID must be a positive integer")
		return
	}
	productID := int32(id64)

	// GET /products/{id}
	if r.Method == http.MethodGet && len(parts) == 2 {
		mu.RLock()
		p, ok := products[productID]
		mu.RUnlock()
		if !ok {
			writeErr(w, http.StatusNotFound, "NOT_FOUND", "Product not found", "No product exists with the given productId")
			return
		}
		writeJSON(w, http.StatusOK, p)
		return
	}

	// POST /products/{id}/details
	if r.Method == http.MethodPost && len(parts) == 3 && parts[2] == "details" {
		var body ProductDetails
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&body); err != nil {
			writeErr(w, http.StatusBadRequest, "INVALID_INPUT", "The provided input data is invalid", "Request body must be valid JSON and match schema")
			return
		}

		// minimal validation (你之后按 api.yaml 再加强)
		if body.SKU == "" || body.Manufacturer == "" || body.CategoryID <= 0 || body.Weight <= 0 {
			writeErr(w, http.StatusBadRequest, "INVALID_INPUT", "The provided input data is invalid", "Missing or invalid required fields")
			return
		}

		// 强制 product_id 和 path 一致（常见要求）
		body.ProductID = productID

		mu.Lock()
		products[productID] = body
		mu.Unlock()

		// 这里先用 200；如果 spec 写 201 再改
		writeJSON(w, http.StatusOK, body)
		return
	}

	writeErr(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", "Check HTTP method and path")
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
