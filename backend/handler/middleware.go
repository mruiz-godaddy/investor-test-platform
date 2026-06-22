package handler

import (
	"log"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"backend/model"
	"backend/store"
)

// ResolveShopper resolves the shopper from the request.
// Resolution order: 1. query param, 2. JWT header, 3. error.
func ResolveShopper(r *http.Request, s *store.Store) (*model.Shopper, error) {
	// 1. Query param
	shopperID := r.URL.Query().Get("shopperId")
	if shopperID != "" {
		return s.GetOrCreateShopper(shopperID)
	}

	// 2. JWT header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		extracted, err := extractShopperFromJWT(authHeader)
		if err == nil && extracted != "" {
			log.Printf("RESOLVED shopper=%s path=%s", extracted, r.URL.Path)
			return s.GetOrCreateShopper(extracted)
		}
	}

	// 3. No shopper
	return nil, fmt.Errorf("no shopperId in query param or Authorization header")
}

func extractShopperFromJWT(authHeader string) (string, error) {
	// Strip "Bearer " or "sso-jwt " prefix
	token := authHeader
	if strings.HasPrefix(token, "Bearer ") {
		token = strings.TrimPrefix(token, "Bearer ")
	} else if strings.HasPrefix(token, "sso-jwt ") {
		token = strings.TrimPrefix(token, "sso-jwt ")
	}
	// Split into parts
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid JWT format")
	}
	// Base64url-decode the payload (part[1])
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", err
	}
	// Parse JSON, extract shopperId
	var claims struct {
		ShopperID string `json:"shopperId"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", err
	}
	if claims.ShopperID == "" {
		return "", fmt.Errorf("no shopperId claim in JWT")
	}
	return claims.ShopperID, nil
}

// WriteError writes a JSON error response with top-level code and message fields.
func WriteError(w http.ResponseWriter, code string, message string, httpStatus int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(map[string]string{
		"code":    code,
		"message": message,
	})
}

// WriteJSON writes a JSON response with the given status code.
func WriteJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
