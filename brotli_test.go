package brotli

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"

	"gopkg.in/kothar/brotli-go.v0/dec"
	"gopkg.in/kothar/brotli-go.v0/enc"
)

func TestSimpleString(T *testing.T) {
	testCompress([]byte("Hello Hello Hello, Hello Hello Hello"), T)
}

func TestShortString(T *testing.T) {
	s := []byte("The quick brown fox")
	l := len(s)

	// Brotli will not compress arrays shorter than 3 characters
	for ; l > 3; l-- {
		testCompress(s[:l], T)
	}
}

func testCompress(s []byte, T *testing.T) {
	T.Logf("Compressing: %s\n", s)

	params := enc.NewBrotliParams()
	buffer1 := make([]byte, len(s)*2)
	encoded, cerr := enc.CompressBuffer(params, s, buffer1)
	if cerr != nil {
		T.Error(cerr)
	}

	buffer2 := make([]byte, len(s))
	decoded, derr := dec.DecompressBuffer(encoded, buffer2)
	if derr != nil {
		T.Error(derr)
	}

	if !bytes.Equal(s, decoded) {
		T.Logf("Decompressed: %s\n", decoded)
		T.Error("Decoded output does not match original input")
	}
}

// Run roundtrip tests from Brotli repository
func TestRoundtrip(T *testing.T) {
	inputs := []string{
		"testdata/alice29.txt",
		"testdata/asyoulik.txt",
		"testdata/lcet10.txt",
		"testdata/plrabn12.txt",
		"enc/encode.cc",
		"shared/dictionary.h",
		"dec/decode.c",
	}

	for _, file := range inputs {
		var err error
		var input []byte

		input, err = ioutil.ReadFile(file)
		if err != nil {
			T.Error(err)
		}

		for _, quality := range []int{1, 6, 9, 11} {
			T.Logf("Roundtrip testing %s at quality %d", file, quality)

			params := enc.NewBrotliParams()
			params.SetQuality(quality)

			bro := testCompressBuffer(params, input, T)

			testDecompressBuffer(input, bro, T)

			testDecompressStream(input, bytes.NewReader(bro), T)

			// Stream compress
			buffer := new(bytes.Buffer)
			testCompressStream(params, input, buffer, T)

			testDecompressBuffer(input, buffer.Bytes(), T)

			// Stream roundtrip
			reader, writer := io.Pipe()
			go testCompressStream(params, input, writer, T)
			testDecompressStream(input, reader, T)
		}
	}
}

// Run roundtrip with a custom dictionary
func TestRoundtripDict(T *testing.T) {
	inputs := []string{
		"testdata/alice29.txt",
		"testdata/asyoulik.txt",
		"testdata/lcet10.txt",
		"testdata/plrabn12.txt",
		"enc/encode.cc",
		"shared/dictionary.h",
		"dec/decode.c",
	}

	dict := []byte("was beginning to get very tired of sitting by her")

	for _, file := range inputs {
		var err error
		var input []byte

		input, err = ioutil.ReadFile(file)
		if err != nil {
			T.Error(err)
		}

		for _, quality := range []int{1, 6, 9, 11} {
			T.Logf("Roundtrip testing %s at quality %d", file, quality)

			params := enc.NewBrotliParams()
			params.SetQuality(quality)

			bro := testCompressBufferDict(params, input, dict, T)

			testDecompressBufferDict(input, bro, dict, T)
		}
	}
}

func testCompressBuffer(params *enc.BrotliParams, input []byte, T *testing.T) []byte {
	// Test buffer compression
	bro, err := enc.CompressBuffer(params, input, nil)
	if err != nil {
		T.Error(err)
	}
	T.Logf("  Compressed from %d to %d bytes, %.1f%%", len(input), len(bro), (float32(len(bro))/float32(len(input)))*100)

	return bro
}

func testDecompressBuffer(input, bro []byte, T *testing.T) {
	// Buffer decompression
	unbro, err := dec.DecompressBuffer(bro, nil)
	if err != nil {
		T.Error(err)
	}

	check("Buffer decompress", input, unbro, T)
}

func testCompressBufferDict(params *enc.BrotliParams, input []byte, inputDict []byte, T *testing.T) []byte {
	// Test buffer compression
	bro, err := enc.CompressBufferDict(params, input, inputDict, nil)
	if err != nil {
		T.Error(err)
	}
	T.Logf("  Compressed from %d to %d bytes, %.1f%%", len(input), len(bro), (float32(len(bro))/float32(len(input)))*100)

	return bro
}

func testDecompressBufferDict(input, bro []byte, inputDict []byte, T *testing.T) {
	// Buffer decompression
	unbro, err := dec.DecompressBufferDict(bro, inputDict, nil)
	if err != nil {
		T.Error(err)
	}

	check("Buffer decompress", input, unbro, T)
}

func testDecompressStream(input []byte, reader io.Reader, T *testing.T) {
	// Stream decompression - use ridiculously small buffer on purpose to
	// test NEEDS_MORE_INPUT state, cf. https://github.com/kothar/brotli-go/issues/28
	streamUnbro, err := ioutil.ReadAll(dec.NewBrotliReaderSize(reader, 128))
	if err != nil {
		T.Error(err)
	}

	check("Stream decompress", input, streamUnbro, T)
}

func testCompressStream(params *enc.BrotliParams, input []byte, writer io.Writer, T *testing.T) {
	bwriter := enc.NewBrotliWriter(params, writer)
	n, err := bwriter.Write(input)
	if err != nil {
		T.Error(err)
	}

	err = bwriter.Close()
	if err != nil {
		T.Error(err)
	}

	if n != len(input) {
		T.Error("Not all input was consumed")
	}
}

func check(test string, input, output []byte, T *testing.T) {
	if len(input) != len(output) {
		T.Errorf("  %s: Length of decompressed output (%d) doesn't match input (%d)", test, len(output), len(input))
	}

	if !bytes.Equal(input, output) {
		T.Errorf("  %s: Input does not match decompressed output", test)
	}
}
