package brotli

import (
	"bytes"
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

		input, err := ioutil.ReadFile(file)
		if err != nil {
			T.Error(err)
		}

		for _, quality := range []int{1, 6, 9, 11} {
			T.Logf("Roundtrip testing %s at quality %d", file, quality)

			params := enc.NewBrotliParams()
			params.SetQuality(quality)

			bro, err := enc.CompressBuffer(params, input, nil)
			if err != nil {
				T.Error(err)
			}
			T.Logf("  Compressed from %d to %d bytes, %.1f%%", len(input), len(bro), (float32(len(bro))/float32(len(input)))*100)

			unbro, err := dec.DecompressBuffer(bro, nil)
			if err != nil {
				T.Error(err)
			}

			if len(input) != len(unbro) {
				T.Errorf("Length of decompressed output (%d) doesn't match input (%d)", len(unbro), len(input))
			}

			if !bytes.Equal(input, unbro) {
				T.Error("  Input does not match decompressed output")
			}
		}
	}
}

func cap(s []byte, length int) []byte {
	if len(s) > length {
		return s[:length]
	} else {
		return s
	}
}
