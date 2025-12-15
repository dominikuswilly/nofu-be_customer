package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		origin         string
		expectedOrigin string
		method         string
		expectedStatus int
	}{
		{
			name:           "Allowed Origin Localhost",
			origin:         "http://localhost:8080",
			expectedOrigin: "http://localhost:8080",
			method:         "GET",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Allowed Origin Cloud",
			origin:         "https://app.netbird.cloud:8090",
			expectedOrigin: "https://app.netbird.cloud:8090",
			method:         "GET",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Allowed Origin Cloud HTTP",
			origin:         "http://app.netbird.cloud:8090",
			expectedOrigin: "http://app.netbird.cloud:8090",
			method:         "GET",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Disallowed Origin",
			origin:         "http://example.com",
			expectedOrigin: "",
			method:         "GET",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Preflight Request Allowed",
			origin:         "http://localhost:8080",
			expectedOrigin: "http://localhost:8080",
			method:         "OPTIONS",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, "/", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Origin", tt.origin)

			rr := httptest.NewRecorder()
			handler := CORSMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			if gotOrigin := rr.Header().Get("Access-Control-Allow-Origin"); gotOrigin != tt.expectedOrigin {
				t.Errorf("handler returned wrong Access-Control-Allow-Origin: got %v want %v",
					gotOrigin, tt.expectedOrigin)
			}
		})
	}
}
