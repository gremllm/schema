package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGremllmMiddleware_PassThrough(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<html><body><h1>Hello</h1></body></html>"))
	})

	wrapped := GremllmMiddleware(handler)
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	// Without ?gremllm, should pass through unchanged
	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "<html>") {
		t.Error("Should return HTML without ?gremllm")
	}
}

func TestGremllmMiddleware_ConvertsToMarkdown(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<html><body><h1>Title</h1><p>Content</p></body></html>"))
	})

	wrapped := GremllmMiddleware(handler)
	req := httptest.NewRequest("GET", "/test?gremllm", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/markdown") {
		t.Errorf("Expected text/markdown, got %s", contentType)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "# Title") {
		t.Errorf("Expected markdown heading, got: %s", body)
	}
	if !strings.Contains(body, "Content") {
		t.Error("Expected content preserved")
	}
}

func TestGremllmMiddleware_NonHTMLPassThrough(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"key": "value"}`))
	})

	wrapped := GremllmMiddleware(handler)
	req := httptest.NewRequest("GET", "/test?gremllm", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	// Non-HTML should pass through unchanged even with ?gremllm
	body := rec.Body.String()
	if body != `{"key": "value"}` {
		t.Errorf("JSON should pass through unchanged, got: %s", body)
	}
}

func TestGremllmMiddleware_ErrorPassThrough(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("<html><body>Not Found</body></html>"))
	})

	wrapped := GremllmMiddleware(handler)
	req := httptest.NewRequest("GET", "/test?gremllm", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	// Error responses should pass through unchanged
	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", rec.Code)
	}
}

func TestGremllmMiddleware_Cache(t *testing.T) {
	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<html><body><h1>Cached</h1></body></html>"))
	})

	wrapped := GremllmMiddleware(handler)

	// First request
	req1 := httptest.NewRequest("GET", "/test?gremllm", nil)
	rec1 := httptest.NewRecorder()
	wrapped.ServeHTTP(rec1, req1)
	body1 := rec1.Body.String()

	// Second request with same content
	req2 := httptest.NewRequest("GET", "/test?gremllm", nil)
	rec2 := httptest.NewRecorder()
	wrapped.ServeHTTP(rec2, req2)
	body2 := rec2.Body.String()

	// Both should produce same result
	if body1 != body2 {
		t.Error("Cached response should be identical")
	}
	if !strings.Contains(body1, "# Cached") {
		t.Error("Should contain converted markdown")
	}
}

func TestGremllmMiddleware_StripsNavFooter(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body>
			<nav>Navigation</nav>
			<main><h1>Main Content</h1></main>
			<footer>Footer</footer>
		</body></html>`))
	})

	wrapped := GremllmMiddleware(handler)
	req := httptest.NewRequest("GET", "/test?gremllm", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	body := rec.Body.String()
	if strings.Contains(body, "Navigation") {
		t.Error("Should strip nav content")
	}
	if strings.Contains(body, "Footer") {
		t.Error("Should strip footer content")
	}
	if !strings.Contains(body, "Main Content") {
		t.Error("Should preserve main content")
	}
}

func TestGremllmMiddleware_EmptyResponse(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		// Empty body
	})

	wrapped := GremllmMiddleware(handler)
	req := httptest.NewRequest("GET", "/test?gremllm", nil)
	rec := httptest.NewRecorder()

	// Should not panic
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestGremllmMiddleware_LargeResponse(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<html><body>"))
		for i := 0; i < 1000; i++ {
			w.Write([]byte("<p>Paragraph content here</p>"))
		}
		w.Write([]byte("</body></html>"))
	})

	wrapped := GremllmMiddleware(handler)
	req := httptest.NewRequest("GET", "/test?gremllm", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if len(body) == 0 {
		t.Error("Should produce output for large response")
	}
}

func TestHashContent(t *testing.T) {
	// Same content should produce same hash
	content := []byte("test content")
	hash1 := hashContent(content)
	hash2 := hashContent(content)

	if hash1 != hash2 {
		t.Error("Same content should produce same hash")
	}

	// Different content should produce different hash
	hash3 := hashContent([]byte("different content"))
	if hash1 == hash3 {
		t.Error("Different content should produce different hash")
	}

	// Hash should be hex string of expected length (MD5 = 32 chars)
	if len(hash1) != 32 {
		t.Errorf("Hash should be 32 chars, got %d", len(hash1))
	}
}

func TestCopyHeaders(t *testing.T) {
	src := http.Header{}
	src.Set("Content-Type", "text/html")
	src.Set("X-Custom", "value")

	dst := http.Header{}
	copyHeaders(dst, src)

	if dst.Get("Content-Type") != "text/html" {
		t.Error("Content-Type not copied")
	}
	if dst.Get("X-Custom") != "value" {
		t.Error("X-Custom not copied")
	}
}

// Benchmark test
func BenchmarkGremllmMiddleware(b *testing.B) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html>
<html>
<body>
	<nav>Nav</nav>
	<main>
		<h1>Title</h1>
		<p>Some <strong>important</strong> content here.</p>
		<ul><li>Item 1</li><li>Item 2</li></ul>
	</main>
	<footer>Footer</footer>
</body>
</html>`))
	})

	wrapped := GremllmMiddleware(handler)

	// Clear cache for accurate benchmark
	cacheMu.Lock()
	cache = make(map[string]cacheEntry)
	cacheMu.Unlock()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test?gremllm", nil)
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)
		io.Copy(io.Discard, rec.Body)
	}
}

func BenchmarkGremllmMiddleware_Cached(b *testing.B) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html>
<html>
<body>
	<h1>Title</h1>
	<p>Content</p>
</body>
</html>`))
	})

	wrapped := GremllmMiddleware(handler)

	// Prime the cache
	req := httptest.NewRequest("GET", "/test?gremllm", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test?gremllm", nil)
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)
		io.Copy(io.Discard, rec.Body)
	}
}

