// cors.go

package middleware

import (
	"log"
	"net/http"
	"strings"
)

func CORSMiddleware(next http.Handler) http.Handler {
	allowedOrigins := map[string]bool{
		"http://localhost:8080":          true,
		"http://localhost:3000":          true,
		"http://100.80.129.244:8080":     true,
		"https://app.netbird.cloud:8090": true,
		"http://app.netbird.cloud:8090":  true,
		// Add this for your request's origin
		"http://app.netbird.cloud:8080":  true,
		"https://app.netbird.cloud:8080": true,
		// Or for flexibility: allow any *.netbird.cloud (but restrict in prod)
		"http://100.80.201.235": true,
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		isAllowed := false
		if origin != "" {
			// Check exact match first
			if allowedOrigins[origin] {
				isAllowed = true
			} else {
				// Optional: Wildcard check (e.g., for subdomains)
				for allowed := range allowedOrigins {
					if strings.HasSuffix(allowed, "*") && strings.HasPrefix(origin, strings.TrimSuffix(allowed, "*")) {
						isAllowed = true
						break
					}
				}
			}
		}

		// Set CORS headers if allowed (or for simple requests)
		if isAllowed || origin == "" { // Allow empty origin for non-browser tests
			log.Printf("✅ Allowed origin: %s", origin)
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "86400") // Cache preflight for 24h
		} else if origin != "" {
			log.Printf("❌ CORS blocked request from origin: %s", origin)
			// For blocked OPTIONS, return error immediately
			if r.Method == "OPTIONS" {
				w.Header().Set("Content-Type", "application/json")
				http.Error(w, `{"error": "CORS origin not allowed"}`, http.StatusForbidden)
				return
			}
		}

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, ...")
			if !isAllowed && origin != "" {
				http.Error(w, `{"error": "CORS origin not allowed"}`, http.StatusForbidden)
			} else {
				w.WriteHeader(http.StatusOK)
			}
			return
		}

		next.ServeHTTP(w, r)
	})
}
