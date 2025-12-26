package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"unsafe"

	"github.com/gremllm/schema/internal/converter"
)

//export Convert
func Convert(htmlInput *C.char, stripNav C.int, stripAside C.int, stripScript C.int) *C.char {
	if htmlInput == nil {
		return C.CString("")
	}

	// Convert C string to Go string
	goHTML := C.GoString(htmlInput)

	// Use the converter package to process HTML with options
	// Convert C ints to Go bools
	stripConfig := converter.StripConfig{
		StripNav:    stripNav != 0,
		StripAside:  stripAside != 0,
		StripScript: stripScript != 0,
	}
	processed, err := converter.ProcessHTML([]byte(goHTML), stripConfig)
	if err != nil {
		return C.CString(goHTML)
	}

	return C.CString(string(processed))
}

//export Free
func Free(str *C.char) {
	C.free(unsafe.Pointer(str))
}

func main() {}
