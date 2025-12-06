package middleware

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"time"
)

type LoggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func NewLoggingResponseWriter(w http.ResponseWriter) *LoggingResponseWriter {
	return &LoggingResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK, // default status
		body:           bytes.NewBuffer([]byte{}),
	}
}

func (lrw *LoggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *LoggingResponseWriter) Write(b []byte) (int, error) {
	lrw.body.Write(b) // capture response body
	return lrw.ResponseWriter.Write(b)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// --- Capture request body ---
		var reqBody []byte
		if r.Body != nil {
			reqBody, _ = io.ReadAll(r.Body)
		}
		r.Body = io.NopCloser(bytes.NewBuffer(reqBody))

		// --- Log Request ---
		log.Println("========== REQUEST ==========")
		log.Printf("➡️ %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		log.Printf("Headers: %v", r.Header)
		log.Printf("Body: %s", string(reqBody))

		// --- Capture response ---
		lrw := NewLoggingResponseWriter(w)

		next.ServeHTTP(lrw, r)

		// --- Log Response ---
		log.Println("========== RESPONSE ==========")
		log.Printf("Status: %d", lrw.statusCode)
		log.Printf("Body: %s", lrw.body.String())
		log.Printf("Completed in %v", time.Since(start))
	})
}

func JSONResponse(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set header before calling the next handler
		w.Header().Set("Content-Type", "application/json")

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
