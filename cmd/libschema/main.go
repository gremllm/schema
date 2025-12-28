package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"unsafe"

	"github.com/gremllm/lib/internal/converter"
)

// Convert processes HTML with optional element stripping configuration.
//
// IMPORTANT: The function signature uses **C.char and C.int instead of []*C.char because
// Go slices cannot be properly passed across CGO/FFI boundaries. Go slices require a
// slice header (pointer, length, capacity), but when called from Python/Node.js FFI,
// the caller can only pass a C array pointer. Using a Go slice parameter causes a nil
// pointer dereference because the slice header is not properly initialized.
//
// The correct pattern for passing arrays through CGO is to pass the array pointer
// (as **C.char) and its length (as C.int) separately, then use unsafe.Slice to convert
// to a Go slice inside the function.
//
//export Convert
func Convert(htmlInput *C.char, elementsToStrip **C.char, elementsLen C.int) *C.char {
	if htmlInput == nil {
		return C.CString("")
	}

	// Convert C string to Go string
	goHTML := C.GoString(htmlInput)
	var goElementsToStrip []string

	// Convert C array to Go slice using pointer arithmetic
	if elementsToStrip != nil && elementsLen > 0 {
		// Create a slice from the C array
		cArray := unsafe.Slice(elementsToStrip, elementsLen)
		for _, cstr := range cArray {
			if cstr != nil {
				goElementsToStrip = append(goElementsToStrip, C.GoString(cstr))
			}
		}
	}

	// Use the converter package to process HTML with options
	// Convert C ints to Go bools
	stripConfig := converter.StripConfig{
		ElementsToStrip: goElementsToStrip,
	}
	processed, err := converter.ProcessHTML([]byte(goHTML), stripConfig)
	if err != nil {
		return C.CString(goHTML)
	}

	md, err := converter.HTMLToMarkdown(processed, stripConfig)
	if err != nil {
		return C.CString(goHTML)
	}

	return C.CString(md)
}

//export Free
func Free(str *C.char) {
	C.free(unsafe.Pointer(str))
}

func main() {}
