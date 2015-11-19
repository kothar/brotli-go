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

// don't let the go 1.5.1 GC put its dirty paws in our beautiful C struct (#14)
// cf. https://github.com/youtube/vitess/blob/071d0e649f22034ad4285c7431ac0a2c9c20090d/go/cgzip/zstream.go#L86-L89
type CBrotliState [unsafe.Sizeof(C.BrotliState{})]C.char

// Decompresses a Brotli-encoded stream using the io.Reader interface
type BrotliReader struct {
	reader io.Reader
	state  CBrotliState
	closed bool

	needOutput bool  // State bounces between needing input and output
	err        error // Persistent error

	// Internal buffer for compressed data
	buffer []byte

	availableIn C.size_t
	totalOut    C.size_t
}

// Fill a buffer, p, with the decompressed contents of the stream.
// Returns the number of bytes read, or an error
func (r *BrotliReader) Read(p []byte) (n int, err error) {
	if len(p) == 0 || r.err != nil {
		return 0, r.err
	}

	// Prepare arguments
	maxOutput := len(p)
	availableOut := C.size_t(maxOutput)
	nextOut := (*C.uint8_t)(unsafe.Pointer(&p[0]))

	for availableOut > 0 && r.err == nil {
		// Read more compressed data
		if r.availableIn == 0 && !r.needOutput {
			read, err := r.reader.Read(r.buffer)
			if read > 0 && err == io.EOF {
				err = nil // Let next Read call return (0, io.EOF)
			}
			if err != nil {
				if err == io.EOF {
					err = io.ErrUnexpectedEOF
				}
				r.err = err
			}
			r.availableIn = C.size_t(read)
		}

		if r.availableIn > 0 || r.needOutput {
			// Decompress
			nextIn := (*C.uint8_t)(unsafe.Pointer(&r.buffer[len(r.buffer)-int(r.availableIn)]))
			result := C.BrotliDecompressStream(
				&r.availableIn,
				&r.nextIn,
				&availableOut,
				&nextOut,
				&r.totalOut,
				(*C.BrotliState)(unsafe.Pointer(&r.state[0])),
			)

			n = maxOutput - int(availableOut)
			switch result {
			case C.BROTLI_RESULT_SUCCESS:
				r.err = io.EOF
			case C.BROTLI_RESULT_NEEDS_MORE_OUTPUT:
				r.needOutput = true
				if n > 0 {
					return n, r.err
				}
				r.err = errors.New("Brotli decompression error: needs more output buffer")
			case C.BROTLI_RESULT_ERROR:
				r.err = errors.New("Brotli decompression error")
			case C.BROTLI_RESULT_NEEDS_MORE_INPUT:
				r.needOutput = false
			default:
				r.err = errors.New("Unrecognized Brotli decompression error")
			}
		}
	}
	return n, r.err
}

// Close the reader and clean up any decompressor state.
func (r *BrotliReader) Close() error {
	if r.closed {
		return r.err
	}
	C.BrotliStateCleanup((*C.BrotliState)(unsafe.Pointer(&r.state[0])))
	r.closed = true
	if r.err == nil || r.err == io.EOF {
		r.err = io.ErrClosedPipe // Make sure future operations fail
		return nil
	}
	return r.err
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

	C.BrotliStateInit((*C.BrotliState)(unsafe.Pointer(&r.state[0])))

	runtime.SetFinalizer(r, func(c io.Closer) { c.Close() })

	return r
}
