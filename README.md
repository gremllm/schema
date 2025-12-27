# LLM Schema

[![Tests](https://github.com/gremllm/lib/actions/workflows/test.yml/badge.svg)](https://github.com/gremllm/lib/actions/workflows/test.yml)
[![Release](https://github.com/gremllm/lib/actions/workflows/release.yml/badge.svg)](https://github.com/gremllm/lib/actions/workflows/release.yml)

A Go middleware for serving LLM-optimized versions of web content. Instead of forcing LLMs to parse verbose HTML with navigation, ads, and boilerplate, websites can serve token-minimized markdown versions at `.md` URLs.

## Project Status

**Current Phase:** Basic middleware infrastructure

- ✅ Go 1.25.5 project setup
- ✅ HTTP middleware that intercepts `.md` URLs
- ✅ Path rewriting from `.md` to `.html`
- ✅ Response capture and processing
- ✅ Basic HTML tag stripping (`<header>`, `<footer>`)
- ✅ CGO library for cross-language support
- ✅ Automated cross-platform releases via goreleaser
- ⏳ HTML to Markdown conversion (planned)
- ⏳ Token optimization (planned)

## Project Structure

```
lib/
├── cmd/
│   ├── server/          # Example HTTP server for testing
│   └── libschema/       # CGO library for cross-language support
├── internal/
│   ├── middleware/      # HTTP middleware implementation for go http server
│   └── converter/       # Core HTML processing logic (single source of truth)
├── examples/            # Sample HTML files
├── ffi_tests/           # FFI tests for other languages
├── build/               # CGO build outputs (.so, .dylib, .dll)
└── .github/workflows/   # Automated release workflow
```

## Quick Start

### Run the Example Server

```bash
go run cmd/server/main.go
```

Then visit:
- http://localhost:8080/index.html - Renders HTML normally
- http://localhost:8080/index.md - Returns HTML (will be markdown in future however it should be stripping the header and footer)

### Use the Middleware in Your Project

```go
package main

import (
    "net/http"
    "github.com/gremllm/lib/internal/middleware"
)

func main() {
    // Your existing HTTP handler
    mux := http.NewServeMux()
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("<h1>Hello World</h1>"))
    })

    // Wrap with schema middleware
    handler := middleware.GremllmMiddleware(mux)

    http.ListenAndServe(":8080", handler)
}
```

Now any route can be accessed with `.md` extension:
- `/about.html` → Regular HTML with all tags
- `/about.md` → Optimized HTML with `<header>` and `<footer>` stripped

## How It Works

1. **Request comes in** for `/page.md`
2. **Middleware intercepts** and rewrites path to `/page.html`
3. **Next handler serves** the HTML (your normal route handler)
4. **Middleware captures** the response
5. **Converter processes** the HTML (strips `<header>`, `<footer>`, etc.)
6. **Returns optimized** HTML

## CGO Library for Cross-Language Support

The project includes a CGO-compiled shared library that can be used from any language supporting C FFI.

### Building the Library Locally

```bash
go build -buildmode=c-shared -o build/libschema.so ./cmd/libschema/
```

This generates:
- `build/libschema.so` (or `.dylib` on macOS, `.dll` on Windows)
- `build/libschema.h` - C header file with function signatures

### Using the Library from Python

```python
from ctypes import cdll, c_char_p

lib = cdll.LoadLibrary('./build/libschema.so')
lib.Convert.argtypes = [c_char_p]
lib.Convert.restype = c_char_p

result = lib.Convert(b'<html><header>...</header><main>...</main></html>')
print(result.decode())  # Header stripped, main content preserved
```

See `examples/test_ffi.py` for a complete working example.

### Automated Releases

When you push a tag (e.g., `v0.1.0`), GitHub Actions automatically:
1. Builds CGO binaries for multiple platforms (Linux, macOS, Windows)
2. Creates a GitHub release with all binaries
3. Generates checksums for verification

**To create a release:**
```bash
git tag v0.1.0
git push origin v0.1.0
```

The workflow (`.github/workflows/release.yml`) uses goreleaser to build:
- Linux AMD64/ARM64 (`.so`)
- macOS AMD64/ARM64 (`.dylib`)
- Windows AMD64 (`.dll`)

## Architecture

### Single Source of Truth

All conversion logic lives in `internal/converter`. Both the HTTP middleware and CGO library import and use this package, ensuring consistent behavior across all implementations.

```
internal/converter/     ← Single source of truth
    ├── converter.go    - Core HTML processing logic

internal/middleware/    ← Uses converter
    └── middleware.go   - HTTP middleware wrapper

cmd/libschema/          ← Uses converter
    └── main.go         - CGO exports for FFI
```

## Next Steps

- [ ] Implement HTML to Markdown conversion (in `internal/converter`)
- [ ] Add token optimization strategies
- [ ] Support site-owner annotations (`llm-keep`, `llm-drop` classes)
- [x] Build CGO library for cross-language support
- [ ] Create language-specific wrapper packages (npm, PyPI, NuGet)

## License

See [LICENSE](LICENSE) for details.
