package dec

/*
#include "./decode.h"

uint8_t* decodeBrotliDictionary;
*/
import "C"

import (
	"errors"
	"unsafe"

	"github.com/kothar/brotli-go/shared"
)

func init() {
	// Set up the default dictionary from the data in the shared package
	C.decodeBrotliDictionary = toC(shared.GetDictionary())
}

// Decompress a Brotli-encoded buffer. Uses decodedBuffer as the destination buffer unless it is too small,
// in which case a new buffer is allocated.
// Returns the slice of the decodedBuffer containing the output, or an error.
func DecompressBuffer(encodedBuffer []byte, decodedBuffer []byte) ([]byte, error) {
	// TODO get decoded length
	encodedLength := len(encodedBuffer)

	var decodedSize C.size_t
	success := C.BrotliDecompressedSize(C.size_t(encodedLength), toC(encodedBuffer), &decodedSize)
	if success != 1 {
		// We can't know in advance how much buffer to allocate, so we will just have to guess
		decodedSize = C.size_t(len(encodedBuffer) * 6)
	}

	if len(decodedBuffer) < int(decodedSize) {
		decodedBuffer = make([]byte, decodedSize)
	}

	// The size of the ouput buffer available
	decodedLength := C.size_t(len(decodedBuffer))
	result := C.BrotliDecompressBuffer(C.size_t(encodedLength), toC(encodedBuffer), &decodedLength, toC(decodedBuffer))
	switch result {
	case C.BROTLI_RESULT_SUCCESS:
		// We're finished
		return decodedBuffer[0:decodedLength], nil
	case C.BROTLI_RESULT_NEEDS_MORE_OUTPUT:
		// We needed more output buffer
		decodedBuffer = make([]byte, len(decodedBuffer)*2)
		return DecompressBuffer(encodedBuffer, decodedBuffer)
	case C.BROTLI_RESULT_ERROR:
		return nil, errors.New("Brotli decompression error")
	case C.BROTLI_RESULT_NEEDS_MORE_INPUT:
		// We can't handle streaming more input results here
		return nil, errors.New("Brotli decompression error: needs more input")
	default:
		return nil, errors.New("Unrecognised Brotli decompression error")
	}
}

func toC(array []byte) *C.uint8_t {
	return (*C.uint8_t)(unsafe.Pointer(&array[0]))
}
