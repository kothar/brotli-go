package dec

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"gopkg.in/kothar/brotli-go.v0/enc"
)

func TestStreamDecompression(T *testing.T) {

	input1 := bytes.Repeat([]byte("The quick brown fox jumps over the lazy dog. "), 100000)

	output1 := make([]byte, len(input1)*2)
	params := enc.NewBrotliParams()
	params.SetQuality(4)

	_, err := enc.CompressBuffer(params, input1, output1)
	if err != nil {
		T.Fatal(err)
	}

	// Decompress as a stream
	reader := NewBrotliReader(bytes.NewReader(output1))
	decoded := make([]byte, len(input1))

	read, err := io.ReadFull(reader, decoded)
	if err != nil {
		T.Fatal(err)
	}
	if read != len(input1) {
		T.Errorf("Length of decoded stream (%d) doesn't match input (%d)", read, len(input1))
	}

	T.Logf("Input:  %s", input1[:50])
	T.Logf("Output: %s", decoded[:50])
	if !bytes.Equal(decoded, input1) {
		T.Error("Decoded output does not match original input")
	}

	// Decompress using a shorter buffer
	reader = NewBrotliReader(bytes.NewReader(output1))
	decoded = make([]byte, 500)

	read, err = reader.Read(decoded)
	if err != nil {
		T.Fatal(err)
	}
	if read != len(decoded) {
		T.Errorf("Length of decoded stream (%d) shorter than requested (%d)", read, len(decoded))
	}

	T.Logf("Input:  %s", input1[:50])
	T.Logf("Output: %s", decoded[:50])
	if !bytes.Equal(decoded, input1[:len(decoded)]) {
		T.Error("Decoded output does not match original input")
	}

	// Read next buffer
	read, err = reader.Read(decoded)
	if err != nil {
		T.Fatal(err)
	}
	if read != len(decoded) {
		T.Errorf("Length of decoded stream (%d) shorter than requested (%d)", read, len(decoded))
	}

	T.Logf("Input:  %s", input1[len(decoded):len(decoded)+50])
	T.Logf("Output: %s", decoded[:50])
	if !bytes.Equal(decoded, input1[len(decoded):2*len(decoded)]) {
		T.Error("Decoded output does not match original input")
	}
}

// Attempt to GC error in decoder
func TestGCErrors(T *testing.T) {
	files, err := filepath.Glob("../testdata/*.compressed")
	if err != nil {
		T.Fatal(err)
	}
	decoded := make([]byte, 18123)

	for _, file := range files {
		T.Logf("Decompressing %s\n", file)

		// Decompress stream
		fileReader, err := os.Open(file)
		if err != nil {
			T.Fatal(err)
		}
		reader := NewBrotliReader(fileReader)
		defer reader.Close()

		for {
			read, err := reader.Read(decoded)
			if err != nil {
				if err == io.EOF {
					if read == 0 {
						break
					}
				} else {
					T.Fatal(err)
				}
			}

			// Force garbage collection
			runtime.GC()
		}
	}
}
