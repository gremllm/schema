package middleware

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gremllm/lib/internal/converter"
)

// Cache settings
const (
	maxCacheSize = 1000
	cacheTTL     = 5 * time.Minute
)

// Cache for converted markdown
type cacheEntry struct {
	content   string
	timestamp time.Time
}

var (
	cache      = make(map[string]cacheEntry)
	cacheOrder []string // Track insertion order for LRU eviction
	cacheMu    sync.RWMutex
)

// evictOldest removes n oldest entries from cache (must hold write lock)
func evictOldest(n int) {
	if n <= 0 || len(cacheOrder) == 0 {
		return
	}
	if n > len(cacheOrder) {
		n = len(cacheOrder)
	}

	// Remove oldest entries
	for i := 0; i < n; i++ {
		delete(cache, cacheOrder[i])
	}
	cacheOrder = cacheOrder[n:]
}

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

// GremllmMiddleware wraps an existing http.Handler to support ?gremllm query parameter.
// When ?gremllm is present in the URL, captures the response, processes the HTML,
// and returns the cleaned markdown version.
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

			// Check cache first
			htmlBytes := rw.body.Bytes()
			cacheKey := hashContent(htmlBytes)

			cacheMu.RLock()
			entry, found := cache[cacheKey]
			cacheMu.RUnlock()

			var markdown string
			if found && time.Since(entry.timestamp) < cacheTTL {
				markdown = entry.content
			} else {
				// Convert HTML to markdown
				var err error
				markdown, err = converter.HTMLToMarkdown(htmlBytes, converter.StripConfig{})
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				// Cache the result
				cacheMu.Lock()
				// Check if we need to evict
				if len(cache) >= maxCacheSize {
					// Evict oldest entry
					evictOldest(1)
				}

				// Add new entry
				if _, exists := cache[cacheKey]; !exists {
					cacheOrder = append(cacheOrder, cacheKey)
				}
				cache[cacheKey] = cacheEntry{content: markdown, timestamp: time.Now()}
				cacheMu.Unlock()
			}

			// Return the processed markdown
			w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
			w.WriteHeader(rw.statusCode)
			w.Write([]byte(markdown))
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

// hashContent creates a cache key from content
func hashContent(content []byte) string {
	h := md5.Sum(content)
	return hex.EncodeToString(h[:])
}
