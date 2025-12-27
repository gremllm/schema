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
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		body:           &bytes.Buffer{},
		statusCode:     http.StatusOK,
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.body.Write(b)
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
}

// GremllmMiddleware wraps an existing http.Handler to support .md URLs
// When a .md URL is requested, it rewrites to .html, captures the response,
// and returns it (for now as-is, later will convert to markdown)
func GremllmMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Check if this is a .md request
		if strings.HasSuffix(path, ".md") {
			// Rewrite the path to .html
			htmlPath := strings.TrimSuffix(path, ".md") + ".html"

			// Special case: index.html should be served as /
			// to avoid FileServer redirects
			if strings.HasSuffix(htmlPath, "/index.html") {
				htmlPath = strings.TrimSuffix(htmlPath, "index.html")
			}

			r.URL.Path = htmlPath

			// Capture the response
			rw := newResponseWriter(w)

			// Call the next handler (which will serve the HTML)
			next.ServeHTTP(rw, r)

			// Process the HTML: strip header and footer tags using converter
			processed, err := converter.ProcessHTML(rw.body.Bytes(), converter.StripConfig{StripNav: true, StripAside: true, StripScript: true})
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
			// Not a .md request, just pass through
			next.ServeHTTP(w, r)
		}
	})
}
