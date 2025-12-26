#!/bin/bash
set -e

cat << 'EOF' > RELEASE_NOTES.md
## LLM Schema CGO Binaries

This release includes cross-platform CGO shared libraries that can be used from any language supporting C FFI.

### Available Libraries:
- Linux AMD64 (`.so`)
- Linux ARM64 (`.so`)
- macOS AMD64 (`.dylib`)
- macOS ARM64 (`.dylib`)
- Windows AMD64 (`.dll`)

### Usage Examples

**Python:**
```python
EOF

cat ffi_tests/python/main.py >> RELEASE_NOTES.md

cat << 'EOF' >> RELEASE_NOTES.md
```

**JavaScript (Node.js):**
```javascript
EOF

cat ffi_tests/nodejs/main.js >> RELEASE_NOTES.md

cat << 'EOF' >> RELEASE_NOTES.md
```
EOF

echo "Release notes generated successfully"
