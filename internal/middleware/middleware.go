package middleware

import (
	"bytes"
	"net/http"
	"strings"

	"github.com/gremllm/lib/internal/converter"
)

// responseWriter is a wrapper around http.ResponseWriter that captures the response
type responseWriter struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
	headers    http.Header
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		body:           &bytes.Buffer{},
		statusCode:     http.StatusOK,
		headers:        make(http.Header),
	}
}

func (rw *responseWriter) Header() http.Header {
	return rw.headers
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.body.Write(b)
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
}

// GremllmMiddleware wraps an existing http.Handler to support ?gremllm query parameter
// When ?gremllm is present in the URL, captures the response, processes the HTML,
// and returns the cleaned version
func GremllmMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if ?gremllm query parameter is present
		_, hasGremllm := r.URL.Query()["gremllm"]

		if hasGremllm {
			// Capture the response
			rw := newResponseWriter(w)

			// Call the next handler (which will serve the HTML)
			next.ServeHTTP(rw, r)

			// Only process successful HTML responses
			if rw.statusCode != http.StatusOK {
				// Pass through non-200 responses unchanged
				copyHeaders(w.Header(), rw.headers)
				w.WriteHeader(rw.statusCode)
				w.Write(rw.body.Bytes())
				return
			}

			contentType := rw.headers.Get("Content-Type")
			if !strings.HasPrefix(contentType, "text/html") {
				// Pass through non-HTML responses unchanged
				copyHeaders(w.Header(), rw.headers)
				w.WriteHeader(rw.statusCode)
				w.Write(rw.body.Bytes())
				return
			}

			// Process the HTML
			processed, err := converter.ProcessHTML(rw.body.Bytes(), converter.StripConfig{})
			if err != nil {
				// If processing fails, return the original HTML
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Return the processed HTML
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(rw.statusCode)
			w.Write(processed)
		} else {
			// No ?gremllm parameter, just pass through
			next.ServeHTTP(w, r)
		}
	})
}

// copyHeaders copies headers from src to dst
func copyHeaders(dst, src http.Header) {
	for k, v := range src {
		dst[k] = v
	}
}
