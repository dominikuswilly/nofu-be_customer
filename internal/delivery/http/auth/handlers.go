package auth

import (
	"be_customer/internal/middleware"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type AuthHandler struct{}

func NewAuthHandler(r *mux.Router) {
	handler := &AuthHandler{}
	api := r.PathPrefix("/api/customer/auth").Subrouter()
	api.HandleFunc("/validate", handler.Validate).Methods(http.MethodPost)
}

func (h *AuthHandler) Validate(w http.ResponseWriter, r *http.Request) {
	// Expect token in Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	bearerToken := strings.Split(authHeader, " ")
	if len(bearerToken) != 2 || strings.ToLower(bearerToken[0]) != "bearer" {
		http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
		return
	}

	tokenString := bearerToken[1]

	_, claims, err := middleware.ValidateToken(tokenString)
	if err != nil {
		statusCode := http.StatusUnauthorized
		if strings.Contains(err.Error(), "server misconfiguration") {
			statusCode = http.StatusInternalServerError
		}
		http.Error(w, "Invalid token: "+err.Error(), statusCode)
		return
	}

	// Token is valid, return claims
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":  true,
		"claims": claims,
	})
}
