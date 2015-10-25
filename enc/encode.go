package enc

/*
// Parts of original C++ header
#include "./encode_go.h"

uint8_t* kBrotliDictionary;
*/
import "C"

import (
	"errors"
	"unsafe"

	"github.com/kothar/brotli-go/shared"
)

func init() {
	// Set up the default dictionary from the data in the shared package
	C.kBrotliDictionary = toC(shared.GetDictionary())
}

type Params struct {
	c C.struct_BrotliParams
}

// Instantiates the compressor parameters with the default settings
func NewParams() *Params {
	params := &Params{C.struct_BrotliParams{
		mode:                    C.MODE_GENERIC,
		quality:                 11,
		lgwin:                   22,
		lgblock:                 0,
		enable_dictionary:       true,
		enable_transforms:       false,
		greedy_block_split:      false,
		enable_context_modeling: true,
	}}

	return params
}

/* Compress a buffer. Uses encodedBuffer as the destination buffer unless it is too small,
 * in which case a new buffer is allocated.
 * Default parameters are used if params is nil.
 * Returns the slice of the encodedBuffer containing the output, or an error.
 */
func CompressBuffer(params *Params, inputBuffer []byte, encodedBuffer []byte) ([]byte, error) {
	inputLength := len(inputBuffer)
	// TODO determine maximum block overhead needed
	if len(encodedBuffer) < inputLength+100 {
		encodedBuffer = make([]byte, inputLength+100)
	}
	if params == nil {
		params = NewParams()
	}

	encodedLength := C.size_t(len(encodedBuffer))
	result := C.BrotliCompressBuffer(params.c, C.size_t(inputLength), toC(inputBuffer), &encodedLength, toC(encodedBuffer))
	if result == 0 {
		return nil, errors.New("Brotli compression error")
	}
	return encodedBuffer[0:encodedLength], nil
}

func toC(array []byte) *C.uint8_t {
	return (*C.uint8_t)(unsafe.Pointer(&array[0]))
}
