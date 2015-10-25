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

/* Decompress a Brotli-encoded buffer. Uses decodedBuffer as the destination buffer unless it is too small,
 * in which case a new buffer is allocated.
 * Returns the slice of the decodedBuffer containing the output, or an error.
 */
func DecompressBuffer(encodedBuffer []byte, decodedBuffer []byte) ([]byte, error) {
	// TODO get decoded length
	encodedLength := len(encodedBuffer)

	var decodedSize C.size_t
	success := C.BrotliDecompressedSize(C.size_t(encodedLength), toC(encodedBuffer), &decodedSize)
	if success != 1 {
		return nil, errors.New("Unable to determine decoded size")
	}

	if len(decodedBuffer) < int(decodedSize) {
		decodedBuffer = make([]byte, decodedSize)
	}

	decodedLength := C.size_t(len(decodedBuffer))
	result := C.BrotliDecompressBuffer(C.size_t(encodedLength), toC(encodedBuffer), &decodedLength, toC(decodedBuffer))
	if result != C.BROTLI_RESULT_SUCCESS {
		// We can't handle streaming more input/output results here
		return nil, errors.New("Brotli decompression error")
	}

	return decodedBuffer[0:decodedLength], nil
}

func toC(array []byte) *C.uint8_t {
	return (*C.uint8_t)(unsafe.Pointer(&array[0]))
}
