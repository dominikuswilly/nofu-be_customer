package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const ContextUserKey contextKey = "user"

func ValidateToken(tokenString string) (*jwt.Token, jwt.MapClaims, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, nil, fmt.Errorf("server misconfiguration: JWT_SECRET not set")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Verify issuer if set in environment
		expectedIssuer := os.Getenv("JWT_ISSUER")
		if expectedIssuer == "" {
			expectedIssuer = "nofu-customer-api" // Default used in handlers.go
		}

		if iss, ok := claims["iss"].(string); ok {
			if iss != expectedIssuer {
				return nil, nil, fmt.Errorf("invalid token issuer")
			}
		}
		return token, claims, nil
	}

	return nil, nil, fmt.Errorf("invalid token claims")
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		_, claims, err := ValidateToken(tokenString)
		if err != nil {
			statusCode := http.StatusUnauthorized
			if strings.Contains(err.Error(), "server misconfiguration") {
				statusCode = http.StatusInternalServerError
			}
			http.Error(w, "Invalid token: "+err.Error(), statusCode)
			return
		}

		// Add claims to context
		ctx := context.WithValue(r.Context(), ContextUserKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
