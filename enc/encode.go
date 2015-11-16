// Brotli compression library bindings for the encoder
package enc

/*
// Parts of original C++ header
#include "./encode_go.h"

typedef uint8_t dict[122784];
dict* kBrotliDictionary;

// Based on BrotliCompressBufferParallel
// https://github.com/google/brotli/blob/24469b81d604ddf1976c3e4b633523bd8f6f631c/enc/encode_parallel.cc#L233
size_t BrotliMaxOutputSize(CBrotliParams params, size_t input_size) {
  // Sanitize params.
  if (params.lgwin < kMinWindowBits) {
    params.lgwin = kMinWindowBits;
  } else if (params.lgwin > kMaxWindowBits) {
    params.lgwin = kMaxWindowBits;
  }
  if (params.lgblock == 0) {
    params.lgblock = 16;
    if (params.quality >= 9 && params.lgwin > params.lgblock) {
      params.lgblock = params.lgwin < 21 ? params.lgwin : 21;
    }
  } else if (params.lgblock < kMinInputBlockBits) {
    params.lgblock = kMinInputBlockBits;
  } else if (params.lgblock > kMaxInputBlockBits) {
    params.lgblock = kMaxInputBlockBits;
  }

  size_t input_block_size = 1 << params.lgblock;
  size_t output_block_size = input_block_size + (input_block_size >> 3) + 1024;

  size_t blocks = (input_size / input_block_size) + 1;

  size_t max_output_size = blocks * output_block_size;

  return max_output_size;
}
*/
import "C"

import (
	"errors"
	"reflect"
	"unsafe"

	"gopkg.in/kothar/brotli-go.v0/shared"
)

func init() {
	// Set up the default dictionary from the data in the shared package
	C.kBrotliDictionary = (*C.dict)(shared.GetDictionary())
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
	c C.struct_CBrotliParams
}

type BrotliCompressor struct {
	c C.CBrotliCompressor
}

// Instantiates the compressor parameters with the default settings
func NewBrotliParams() *BrotliParams {
	params := &BrotliParams{C.struct_CBrotliParams{
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

// Maximum output size based on
// https://github.com/google/brotli/blob/24469b81d604ddf1976c3e4b633523bd8f6f631c/enc/encode_parallel.cc#L233
// There doesn't appear to be any documentation of what this calculation is based on.
func (p *BrotliParams) MaxOutputSize(inputLength int) int {
	return int(C.BrotliMaxOutputSize(p.c, C.size_t(inputLength)))
}

// Compress a buffer. Uses encodedBuffer as the destination buffer unless it is too small,
// in which case a new buffer is allocated.
// Default parameters are used if params is nil.
// Returns the slice of the encodedBuffer containing the output, or an error.
func CompressBuffer(params *BrotliParams, inputBuffer []byte, encodedBuffer []byte) ([]byte, error) {

	if params == nil {
		params = NewBrotliParams()
	}

	inputLength := len(inputBuffer)
	maxOutSize := params.MaxOutputSize(inputLength)

	if len(encodedBuffer) < maxOutSize {
		encodedBuffer = make([]byte, maxOutSize)
	}

	encodedLength := C.size_t(len(encodedBuffer))
	result := C.CBrotliCompressBuffer(params.c, C.size_t(inputLength), toC(inputBuffer), &encodedLength, toC(encodedBuffer))
	if result == 0 {
		return nil, errors.New("Brotli compression error")
	}
	return encodedBuffer[0:encodedLength], nil
}

func toC(array []byte) *C.uint8_t {
	return (*C.uint8_t)(unsafe.Pointer(&array[0]))
}

func NewBrotliCompressor(params *BrotliParams) *BrotliCompressor {
	if params == nil {
		params = NewBrotliParams()
	}

	return &BrotliCompressor{c: C.CBrotliCompressorNew(params.c)}
}

func (bp *BrotliCompressor) GetInputBlockSize() int64 {
	return int64(C.CBrotliCompressorGetInputBlockSize(bp.c))
}

func (bp *BrotliCompressor) CopyInputToRingBuffer(input []byte) {
	C.CBrotliCompressorCopyInputToRingBuffer(bp.c, C.size_t(len(input)), toC(input))
}

func (bp *BrotliCompressor) WriteBrotliData(isLast bool, forceFlush bool) []byte {
	var outSize C.size_t
	var output *C.uint8_t
	C.CBrotliCompressorWriteBrotliData(bp.c, C.bool(isLast), C.bool(forceFlush), &outSize, &output)
	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(output)),
		Len:  int(outSize),
		Cap:  int(outSize),
	}
	return *(*[]byte)(unsafe.Pointer(&hdr))
}
