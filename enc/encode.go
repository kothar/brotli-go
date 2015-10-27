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

type Mode int

const (
	// Default compression mode. The compressor does not know anything in
	// advance about the properties of the input.
	GENERIC Mode = iota
	// Compression mode for UTF-8 format text input.
	TEXT
	// Compression mode used in WOFF 2.0.
	FONT
)

type BrotliParams struct {
	c C.struct_BrotliParams
}

// Instantiates the compressor parameters with the default settings
func NewBrotliParams() *BrotliParams {
	params := &BrotliParams{C.struct_BrotliParams{
		mode:    C.MODE_GENERIC,
		quality: 11,
		lgwin:   22,
		lgblock: 0,

		// Deprecated according to header
		enable_dictionary:       true,
		enable_transforms:       false,
		greedy_block_split:      false,
		enable_context_modeling: true,
	}}

	return params
}

func (p *BrotliParams) Mode() Mode {
	return Mode(p.c.mode)
}

func (p *BrotliParams) SetMode(value Mode) {
	p.c.mode = C.enum_Mode(value)
}

// Controls the compression-speed vs compression-density tradeoffs. The higher
// the quality, the slower the compression. Range is 0 to 11. Default is 11.
func (p *BrotliParams) Quality() int {
	return int(p.c.quality)
}

func (p *BrotliParams) SetQuality(value int) {
	p.c.quality = C.int(value)
}

// Base 2 logarithm of the sliding window size. Range is 10 to 24. Default is 22.
func (p *BrotliParams) Lgwin() int {
	return int(p.c.lgwin)
}

func (p *BrotliParams) SetLgwin(value int) {
	p.c.lgwin = C.int(value)
}

// Base 2 logarithm of the maximum input block size. Range is 16 to 24.
// If set to 0 (default), the value will be set based on the quality.
func (p *BrotliParams) Lgblock() int {
	return int(p.c.lgblock)
}

func (p *BrotliParams) SetLgblock(value int) {
	p.c.lgblock = C.int(value)
}

// Compress a buffer. Uses encodedBuffer as the destination buffer unless it is too small,
// in which case a new buffer is allocated.
// Default parameters are used if params is nil.
// Returns the slice of the encodedBuffer containing the output, or an error.
func CompressBuffer(params *BrotliParams, inputBuffer []byte, encodedBuffer []byte) ([]byte, error) {
	inputLength := len(inputBuffer)
	// TODO determine maximum block overhead needed
	if len(encodedBuffer) < inputLength+100 {
		encodedBuffer = make([]byte, inputLength+100)
	}
	if params == nil {
		params = NewBrotliParams()
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
