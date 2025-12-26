#!/usr/bin/env python3
"""
Test script to verify the CGO library works via Python FFI
"""

from ctypes import cdll, c_char_p, c_bool
import os

# Load the shared library
lib_path = os.path.join(os.path.dirname(__file__), '..', '..', 'build', 'libschema.so')
lib = cdll.LoadLibrary(lib_path)

# Set up function signatures
lib.Convert.argtypes = [c_char_p, c_bool, c_bool, c_bool]
lib.Convert.restype = c_char_p

# Test HTML
test_html = b"""<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body>
    <header><h1>This should be stripped</h1></header>
    <nav><a href="/">Home</a></nav>
    <main><p>This content should remain</p></main>
    <footer><p>This should also be stripped</p></footer>
</body>
</html>"""

print("Testing Convert()...")
result = lib.Convert(test_html, True, True, True)
result_str = result.decode('utf-8')

print("\nInput HTML:")
print(test_html.decode('utf-8'))
print("\nOutput HTML:")
print(result_str)

failures = 0

# Check that header and footer are stripped
if b'<header>' not in result_str.encode():
    print("\n✓ Header tag successfully stripped")
else:
    print("\n✗ Header tag still present")
    failures += 1

if b'<footer>' not in result_str.encode():
    print("✓ Footer tag successfully stripped")
else:
    print("✗ Footer tag still present")
    failures += 1

if b'<main>' in result_str.encode():
    print("✓ Main content preserved")
else:
    print("✗ Main content missing")
    failures += 1

print("\n" + "="*50)

if failures > 0:
    print(f"{failures} test(s) failed")
    exit(1)

print("CGO library is working correctly!")
