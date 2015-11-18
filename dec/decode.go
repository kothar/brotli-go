// Go wrapper for the Brotli C decoder implementation
package dec

/*
#include "./decode.h"

typedef uint8_t dict[122784];
dict* decodeBrotliDictionary;
*/
import "C"

import (
	"errors"
	"io"
	"runtime"
	"unsafe"

	"gopkg.in/kothar/brotli-go.v0/shared"
)

func init() {
	// Set up the default dictionary from the data in the shared package
	C.decodeBrotliDictionary = (*C.dict)(shared.GetDictionary())
}

// Decompress a Brotli-encoded buffer. Uses decodedBuffer as the destination buffer unless it is too small,
// in which case a new buffer is allocated.
// Returns the slice of the decodedBuffer containing the output, or an error.
func DecompressBuffer(encodedBuffer []byte, decodedBuffer []byte) ([]byte, error) {
	encodedLength := len(encodedBuffer)
	var decodedSize C.size_t

	// If the user has provided a sensibly size buffer, assume they know how long the output should be
	// Otherwise try to determine the correct length from the input
	if len(decodedBuffer) < len(encodedBuffer) {
		success := C.BrotliDecompressedSize(C.size_t(encodedLength), toC(encodedBuffer), &decodedSize)
		if success != 1 {
			// We can't know in advance how much buffer to allocate, so we will just have to guess
			decodedSize = C.size_t(len(encodedBuffer) * 6)
		}

		if len(decodedBuffer) < int(decodedSize) {
			decodedBuffer = make([]byte, decodedSize)
		}
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

// Decompresses a Brotli-encoded stream using the io.Reader interface
type BrotliReader struct {
	state  C.BrotliState
	reader io.Reader

	// Internal buffer for compressed data
	buffer []byte

	availableIn     C.size_t
	nextIn          *C.uint8_t
	totalOut        C.size_t
	eof             bool
	needsMoreOutput bool
}

// Fill a buffer, p, with the decompressed contents of the stream.
// Returns the number of bytes read, or an error
func (r *BrotliReader) Read(p []byte) (int, error) {
	if r.eof {
		return 0, io.EOF
	}

	var err error

	// Prepare arguments
	maxOutput := len(p)
	availableOut := C.size_t(maxOutput)
	nextOut := (*C.uint8_t)(unsafe.Pointer(&p[0]))
	var read int

	for availableOut > 0 {

		// Read more compressed data
		if r.availableIn == 0 {
			var read int
			if read, err = r.reader.Read(r.buffer); err != nil {
				if err == io.EOF {
					r.eof = true
					if !r.needsMoreOutput {
						break
					}
				} else {
					return 0, err
				}
			}

			r.availableIn = C.size_t(read)
			r.nextIn = (*C.uint8_t)(unsafe.Pointer(&r.buffer[0]))
		}

		if r.availableIn > 0 || r.needsMoreOutput {
			r.needsMoreOutput = false
			r.eof = false

			// Decompress
			result := C.BrotliDecompressStream(
				&r.availableIn,
				&r.nextIn,
				&availableOut,
				&nextOut,
				&r.totalOut,
				&r.state,
			)

			read = maxOutput - int(availableOut)

			switch result {
			case C.BROTLI_RESULT_SUCCESS:
				break
			case C.BROTLI_RESULT_NEEDS_MORE_OUTPUT:
				r.needsMoreOutput = true

				if read > 0 {
					return read, nil
				} else {
					return 0, errors.New("Brotli decompression error: needs more output buffer")
				}
			case C.BROTLI_RESULT_ERROR:
				return 0, errors.New("Brotli decompression error")
			case C.BROTLI_RESULT_NEEDS_MORE_INPUT:
				continue
			default:
				return 0, errors.New("Unrecognised Brotli decompression error")
			}
		}
	}

	return read, nil
}

// Close the reader and clean up any decompressor state
func (r *BrotliReader) Close() error {

	C.BrotliStateCleanup(&r.state)

	if v, ok := r.reader.(io.Closer); ok {
		return v.Close()
	}

	return nil
}

// Returns a Reader that decompresses the stream from another reader.
//
// Ensure that you Close the stream when you are finished in order to clean up the
// Brotli decompression state.
//
// The internal decompression buffer defaults to 128kb
func NewBrotliReader(stream io.Reader) *BrotliReader {
	return NewBrotliReaderSize(stream, 128*1024)
}

// The same as NewBrotliReader, but allows the internal buffer size to be set.
//
// The size of the internal buffer may be specified which will hold compressed data
// before being read by the decompressor
func NewBrotliReaderSize(stream io.Reader, size int) *BrotliReader {
	r := &BrotliReader{
		reader: stream,
		buffer: make([]byte, size),
	}

	C.BrotliStateInit(&r.state)

	runtime.SetFinalizer(r, func(c io.Closer) { c.Close() })

	return r
}
